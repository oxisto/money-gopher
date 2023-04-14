<img src="img/gopher.png" width="25%"/>

# The Money Gopher

[![build](https://github.com/oxisto/money-gopher/actions/workflows/build.yml/badge.svg)](https://github.com/oxisto/money-gopher/actions/workflows/build.yml)
[![](https://godoc.org/github.com/oxisto/money-gopher?status.svg)](https://pkg.go.dev/github.com/oxisto/money-gopher)
[![Go Report
Card](https://goreportcard.com/badge/github.com/oxisto/money-gopher)](https://goreportcard.com/report/github.com/oxisto/money-gopher)
[![codecov](https://codecov.io/gh/oxisto/money-gopher/branch/main/graph/badge.svg?token=U2LKZFCGJO)](https://codecov.io/gh/oxisto/money-gopher)


The Money Gopher will help you to keep track of your investments.

# Why?

Surely, there are a number of programs and services out there that already
manage your portfolio(s), why creating another one? Well there are several
reasons or rather requirements that I need. Note, that these might be very
specific to my use case, but maybe somebody else will appreciate them as well.

* 🏘️ I need to manage several portfolios for several distinct people, for
  example my own and my children's. I want to keep these portfolios completely
  separate, but still manageable within the same uniform UI or program. For a
  lack of better term, I call this a "portfolio group" for now.
* 💵 All "portfolio groups" could share stock information, such as buy/sell
  prices and meta-data. Then they only need to be retrieved once and are
  available to all "groups".
* 🤑 Within one "portfolio group", I obviously want to manage several
  portfolios, displaying certain performance values (e.g. absolute gain,
  time-weighted return, etc.) per portfolio and for the whole group.
* 📱 I want to access this information from multiple devices, e.g., my laptop,
  my tablet and my phone. But, I don't necessarily need this information on the
  go, so having some kind of "server" locally to my network and a browser-based
  UI seems to be perfect. This means that the UI tech stack should reflect
  responsiveness and a mobile-friendly design. If I *really* need this
  information on the go, I could then still set this up on a server that I own
  or VPN to my home network.
* 👨‍💻 I love APIs, so having access to this in a RPC or REST API would be
  awesome. It is anyway needed for the UI. Maybe also a simple CLI for quick
  tasks, such as triggering a refresh of stock information would also be nice.

Furthermore, there are some personal technical motivations that drove me to
creating this.

* 📞 I wanted to explore new ways of providing RPC-style APIs that are not based
  on the arguably bloated gRPC framework. Therefore, I am exploring Buf's
  [Connect](https://connect.build) framework in this project, which seems
  promising, even for browser-based interactions.
* 🔲 I am still on the spiritual search for a good UI framework, so this might
  be a good chance to explore different options.
* 📈 I wanted to understand the math behind some of the used performance models,
  such as time-weighted rate of return a little bit better.

# When is it finished?

Since I am working on this in my spare time, it will probably take a while 😃.
