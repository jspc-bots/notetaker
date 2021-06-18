package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ergochat/irc-go/ircfmt"
	"github.com/google/go-github/v35/github"
	"github.com/lrstanley/girc"
)

const (
	HelpText = `$bNotetaker Help$r

  $bHELP$r                - This text
  $bNEW [password]$r      - Create a new note, setting a password on the channel.
                        On creation of this channel, you will be invited and further help
                        text will be given (which is also included below).


$iThe following commands only work when given in a notetaker session, and only from the originating user$r


  $bWRITE [session_id]$r              - Write this note to a gist, returning the gist URL
  $bCLOSE [session_id]$r              - Write this note, return gist, then close the notetaker session.
                                    Closing the session, in essence, means that notetaker will leave the
                                    channel and never come back. You will then be able to close it/ delete it.
`
)

type Bot struct {
	client   *girc.Client
	github   *github.Client
	routing  map[*regexp.Regexp]handlerFunc
	sessions map[string]*Session
}

type handlerFunc func(originator string, groups [][]byte) error

func New(user, password, server string, verify bool, gh *github.Client) (b Bot, err error) {
	b.github = gh
	b.sessions = make(map[string]*Session)

	u, err := url.Parse(server)
	if err != nil {
		return
	}

	config := girc.Config{
		Server: u.Hostname(),
		Port:   must(strconv.Atoi(u.Port())).(int),
		Nick:   Nick,
		User:   Nick,
		Name:   Nick,
		SASL: &girc.SASLPlain{
			User: user,
			Pass: password,
		},
		SSL: u.Scheme == "ircs",
		TLSConfig: &tls.Config{
			InsecureSkipVerify: !verify,
		},
	}

	b.client = girc.New(config)
	err = b.addHandlers()

	return
}

func (b *Bot) addHandlers() (err error) {
	b.client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		//c.Cmd.Join(Chan)
	})

	b.routing = make(map[*regexp.Regexp]handlerFunc)

	b.routing[regexp.MustCompile(`(?i)^new\s+(.+)$`)] = b.newNote
	b.routing[regexp.MustCompile(`(?i)^help$`)] = b.getHelp
	b.routing[regexp.MustCompile(`(?i)^save\s+(.+)$`)] = b.saveNote
	b.routing[regexp.MustCompile(`(?i)^close\s+(.+)$`)] = b.closeNote

	// Route messages
	b.client.Handlers.Add(girc.PRIVMSG, b.messageRouter)

	return
}

func (b Bot) messageRouter(c *girc.Client, e girc.Event) {
	var err error

	// skip messages older than a minute (assume it's the replayer)
	cutOff := time.Now().Add(0 - time.Minute)
	if e.Timestamp.Before(cutOff) {
		// ignore
		return
	}

	dst := e.Params[0]

	if session, ok := b.sessions[dst]; ok {
		err = session.process(e)
		if err != nil {
			log.Printf("error processing session bound message: %v", err)
		}

		return
	}

	// From here on, only respond if this is a PRIVMSG directly to us
	if dst != Nick {
		return
	}

	msg := []byte(e.Last())

	for r, f := range b.routing {
		if r.Match(msg) {
			err = f(e.Source.Name, r.FindAllSubmatch(msg, -1)[0])
			if err != nil {
				log.Printf("%v error: %s", f, err)
			}

			return
		}
	}

	// If we get this far, it's a big fat error
	b.client.Cmd.Message(e.Source.Name, "Unknown command. Run HELP to get help")
}

func (b Bot) getHelp(originator string, _ [][]byte) (err error) {
	for _, line := range strings.Split(HelpText, "\n") {
		b.client.Cmd.Message(originator, ircfmt.Unescape(line))
	}

	return
}

func (b Bot) newNote(originator string, groups [][]byte) (err error) {
	// groups[0] is the full string
	var password string
	if len(groups) == 2 {
		password = string(groups[1])
	}

	s, err := NewSession(b.client, b.github, originator, password)
	if err != nil {
		return
	}

	b.sessions[s.channelName] = s

	b.client.Cmd.Messagef(originator, "created channel %s, which you should be invited to join", s.channelName)

	return
}

func (b Bot) saveNote(originator string, groups [][]byte) (err error) {
	id := string(groups[1])

	session, err := b.getValidSession(originator, id)
	if err != nil {
		b.client.Cmd.Message(originator, err.Error())
		b.client.Cmd.Message(session.channelName, err.Error())

		return
	}

	err = session.save()
	if err != nil {
		b.client.Cmd.Message(originator, err.Error())
		b.client.Cmd.Message(session.channelName, err.Error())

		return
	}

	b.client.Cmd.Messagef(originator, "gist location: %s", *session.gist.HTMLURL)
	b.client.Cmd.Messagef(session.channelName, "gist location: %s", *session.gist.HTMLURL)

	return
}

func (b Bot) closeNote(originator string, groups [][]byte) (err error) {
	err = b.saveNote(originator, groups)
	if err != nil {
		return
	}

	id := string(groups[1])

	session, err := b.getValidSession(originator, id)
	if err != nil {
		return
	}

	return session.close()
}

func (b Bot) getValidSession(originator, id string) (s *Session, err error) {
	channelName := fmt.Sprintf("#notetaker-%s", id)

	var ok bool

	if s, ok = b.sessions[channelName]; !ok {
		err = fmt.Errorf("could not find session")

		b.client.Cmd.Message(originator, err.Error())

		return
	}

	if s.user != originator {
		err = fmt.Errorf("you were not the requestor of this session")

		b.client.Cmd.Message(originator, err.Error())
	}

	return
}
