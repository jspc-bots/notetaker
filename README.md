# Notetaker

This bot lives to take notes. It lives in a `/query`/ `/msg` session, and takes down anything you say, storing it in a gist somewhere.

These gists are saved as markdown files, allowing all of the markdown constructs you need.

## Flow

Create a new note

```
> /msg notetaker new [password]
```

Use the password to ensure you're the only author, or to protect anything you don't want to share.

Notetaker will create a new channel, setting the optional password, and then invites the requestor


```irc
20:34 <jspc> HELP
20:34 <notetaker> Notetaker Help
20:34 <notetaker>   HELP                - This text
20:34 <notetaker>   NEW [password]      - Create a new note, setting a password on the channel.
20:34 <notetaker>                         On creation of this channel, you will be invited and further help
20:34 <notetaker>                         text will be given (which is also included below).
20:34 <notetaker> The following commands only work when given in a notetaker session, and only from the originating user
20:34 <notetaker>   WRITE [session_id]              - Write this note to a gist, returning the gist URL
20:34 <notetaker>   CLOSE [session_id]              - Write this note, return gist, then close the notetaker session.
20:34 <notetaker>                                     Closing the session, in essence, means that notetaker will leave the
20:34 <notetaker>                                     channel and never come back. You will then be able to close it/ delete it.
20:34 <jspc> NEW super-secret-password
20:34 <notetaker> created channel #notetaker-c3689g23k1k01ea1odmg, which you should be invited to join
```

From there, all notes are assumed to be in Markdown. Join the channel it invites you to (only invited users can join) and start writing notes. As the notetaker session channel explains:

```irc
21:20 -!- jspc [~jspc@172.17.0.1] has joined #notetaker-c368v0i3k1k1f73f6qhg
21:20 -!- Topic for #notetaker-c368v0i3k1k1f73f6qhg: /MSG notetaker save c368v0i3k1k1f73f6qhg
21:20 -!- Topic set by notetaker [~notetaker@172.17.0.1] [Fri Jun 18 21:20:18 2021]
21:20 [Users #notetaker-c368v0i3k1k1f73f6qhg]
21:20 [@notetaker] [ jspc]
21:20 -!- Irssi: #notetaker-c368v0i3k1k1f73f6qhg: Total of 2 nicks [1 ops, 0 halfops, 0 voices, 1 normal]
21:20 < HistServ> notetaker set channel modes: +is
21:20 < HistServ> notetaker set channel modes: +k super-secret-password
21:20 < HistServ> notetaker set the channel topic to: /MSG notetaker save c368v0i3k1k1f73f6qhg
21:20 <@notetaker> Welcome to Notetaker Session
21:20 <@notetaker> Useful commands:
21:20 <@notetaker>   /MSG notetaker save [id]   - Save these notes. Returns a gist URL
21:20 <@notetaker>   /MSG notetaker close [id]  - Save these notes, then boot the notetaker bot
21:20 <@notetaker> (the ID of this channel is "c368v0i3k1k1f73f6qhg")
21:20 -!- Channel #notetaker-c368v0i3k1k1f73f6qhg created Fri Jun 18 21:20:18 2021
21:20 -!- Irssi: Join to #notetaker-c368v0i3k1k1f73f6qhg was synced in 0 secs
```
