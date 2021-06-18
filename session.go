package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ergochat/irc-go/ircfmt"
	"github.com/google/go-github/v35/github"
	"github.com/lrstanley/girc"
	"github.com/rs/xid"
)

const (
	ChannelWelcomeText = `$bWelcome to Notetaker Session$r
Useful commands:
  $b/MSG notetaker save [id]$r   - Save these notes. Returns a gist URL
  $b/MSG notetaker close [id]$r  - Save these notes, then boot the notetaker bot
`
)

type Session struct {
	gist        *github.Gist
	github      *github.Client
	channelName string
	irc         *girc.Client
	user        string

	lines []string
}

func NewSession(irc *girc.Client, gh *github.Client, user, password string) (s *Session, err error) {
	// Mint an ID
	s = new(Session)

	s.github = gh
	s.irc = irc
	s.user = user

	// Create a channel
	id := xid.New().String()
	s.channelName = fmt.Sprintf("#notetaker-%s", id)

	s.irc.Cmd.Join(s.channelName)
	s.irc.Cmd.Mode(s.channelName, "+is")

	// Set password if we have one
	if password != "" {
		s.irc.Cmd.Mode(s.channelName, "+k", password)
	}

	// Set topic
	s.irc.Cmd.Topic(s.channelName, fmt.Sprintf("/MSG notetaker save %s", id))

	// Write welcome text
	for _, line := range strings.Split(ChannelWelcomeText, "\n") {
		s.irc.Cmd.Message(s.channelName, ircfmt.Unescape(line))
	}

	s.irc.Cmd.Message(s.channelName, ircfmt.Unescape(fmt.Sprintf("(the ID of this channel is $b$u%q$r)", id)))

	// Invite user
	s.irc.Cmd.Invite(s.channelName, s.user)

	return
}

func (s *Session) process(e girc.Event) (err error) {
	msg := e.Last()

	s.storeLine(msg)

	return
}

func (s *Session) storeLine(line string) {
	s.lines = append(s.lines, line)
}

func (s *Session) save() (err error) {
	// If s.gist is set, update that gist
	// otherwise, create new
	var files = map[github.GistFilename]github.GistFile{
		"Notes.md": {
			Filename: github.String("Notes.md"),
			Content:  github.String(strings.Join(s.lines, "\n")),
		},
	}

	if s.gist == nil {
		s.gist, _, err = s.github.Gists.Create(context.Background(), &github.Gist{
			Description: github.String(fmt.Sprintf("Uploaded by notetaker for %s", s.user)),
			Files:       files,
			Public:      github.Bool(false),
		})
	} else {
		s.gist.Files = files
		s.gist, _, err = s.github.Gists.Edit(context.Background(), *s.gist.ID, s.gist)
	}

	return
}

func (s *Session) close() (err error) {
	s.irc.Cmd.Part(s.channelName)

	return
}
