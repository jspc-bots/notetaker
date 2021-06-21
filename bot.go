package main

import (
	"fmt"
	"strings"

	"github.com/ergochat/irc-go/ircfmt"
	"github.com/google/go-github/v35/github"
	"github.com/jspc-bots/bottom"
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
	bottom   bottom.Bottom
	github   *github.Client
	sessions Sessions
}

func New(user, password, server string, verify bool, gh *github.Client) (b Bot, err error) {
	b.github = gh
	b.sessions = make(Sessions)

	b.bottom, err = bottom.New(user, password, server, verify)
	if err != nil {
		return
	}

	router := bottom.NewRouter()
	router.AddRoute(`(?i)^new\s+(.+)$`, b.newNote)
	router.AddRoute(`(?i)^help$`, b.getHelp)
	router.AddRoute(`(?i)^save\s+(.+)$`, b.saveNote)
	router.AddRoute(`(?i)^close\s+(.+)$`, b.closeNote)

	b.bottom.Middlewares.Push(b.sessions)
	b.bottom.Middlewares.Push(router)

	return
}

func (b Bot) getHelp(_, channel string, _ []string) (err error) {
	for _, line := range strings.Split(HelpText, "\n") {
		b.bottom.Client.Cmd.Message(channel, ircfmt.Unescape(line))
	}

	return
}

func (b Bot) newNote(_, channel string, groups []string) (err error) {
	password := groups[1]

	s, err := NewSession(b.bottom.Client, b.github, channel, password)
	if err != nil {
		return
	}

	b.sessions[s.channelName] = s

	b.bottom.Client.Cmd.Messagef(channel, "created channel %s, which you should be invited to join", s.channelName)

	return
}

func (b Bot) saveNote(_, channel string, groups []string) (err error) {
	id := groups[1]

	session, err := b.getValidSession(channel, id)
	if err != nil {
		return
	}

	err = session.save()
	if err != nil {
		return
	}

	b.bottom.Client.Cmd.Messagef(channel, "gist location: %s", *session.gist.HTMLURL)
	b.bottom.Client.Cmd.Messagef(session.channelName, "gist location: %s", *session.gist.HTMLURL)

	return
}

func (b Bot) closeNote(originator, channel string, groups []string) (err error) {
	err = b.saveNote(originator, channel, groups)
	if err != nil {
		return
	}

	id := string(groups[1])

	session, err := b.getValidSession(channel, id)
	if err != nil {
		return
	}

	return session.close()
}

func (b Bot) getValidSession(channel, id string) (s *Session, err error) {
	channelName := fmt.Sprintf("#notetaker-%s", id)

	var ok bool

	if s, ok = b.sessions[channelName]; !ok {
		err = fmt.Errorf("could not find session")

		return
	}

	if s.user != channel {
		err = fmt.Errorf("you were not the requestor of this session")
	}

	return
}
