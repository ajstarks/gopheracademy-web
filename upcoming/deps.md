+++
author = ["Nathan Youngman"]
date = "2014-12-20T08:00:00+00:00"
title = "Managing Dependencies"
series = ["Advent 2014"]
+++

# Packages, dependencies, versions. 

This post will explore two tools and how I've been using them. **Godep** as the consumer of third-party packages, and **gopkg.in** as a library author.

Before getting started, make sure you're comfortable with the fundamentals of [GOPATH][code] and `go get`. If you find GitHub's forking mechanism at odds with the go tool, go read Katrina Owen's excellent [*Contributing to Open Source Git Repositories in Go*][remotes].

While I've found these tools useful, I don't use them for every project. Please read [*The Case Against Third Party Libraries*][case-against-3pl] by Ben Johnson and be mindful of what you need.

# Godep

> "If you're using an externally supplied package and worry that it might change in unexpected ways, the simplest solution is to copy it to your local repository." - [Go FAQ][get_version]

[Godep][] is a tool by Keith Rarick that does this copying for you. 

If you are worried about unexpected changes, Godep can provide a safety net, but then so can backing up your `/src` folder. Personally, this isn't the most compelling reason to use Godep.

Some time ago I put Godep through its paces for [*Go Package Management*][go-packages] and [Packages & Dependencies](https://speakerdeck.com/nathany/go-packages) (slides). Still I had little reason to use it -- not until recently.

While working with a [library for Apple Push Notifications](https://github.com/timehop/apns), I made a few small pull requests. Benny Wong, who maintains the library, was away on his honeymoon at the time. Congrats Benny!

```go
import "github.com/timehop/apns"
```

If my project was a solo effort, I wouldn't feel the need for Godep, but I was working on a team. Suddenly I needed to ensure my team mates had my changes. I wanted something easy to explain while continuing to contribute changes upstream. So I ran:

```console
godep save
```

It copied my current branch from `src/github.com/timehop/apns` into a Godeps folder which I checked in with the project. Then I added instructions in the README to use:

```console
godep go build
```

Doing so will manipulate the GOPATH to use the code copied into myproject/Godeps.

As I made more changes, I found myself working in `src/github.com/timehop/apns` and using `go build` as usual. Before pushing to GitHub, I'd just update the Godeps folder with:

```console
godep update github.com/timehop/apns
```

Godep worked out quite well. With just a few commands, you can snapshot (and restore) your project's dependencies and continue to work as usual.

One downside is that everyone using your project needs Godep. That's fine for a private project, but ideally open source can be installed with just `go get`. Keith Rarick added an `-r` option for this purpose, which the CoreOS team [has had success with][CoreOS].

Now that `-r` is available, there wouldn't be any harm in using Godep for a command-line tool like [Looper](https://github.com/nathany/looper). I just haven't felt the need. Even if a dependency were to disappear, I have the code in `/src`, so I could recover.

Godep isn't something I use for libraries either -- for that, let's take a look at gopkg.in.

# gopkg.in

Versioning is a crutch. 

If I could design a solid, stable API from the beginning, would I bother with version numbers? Nope.

Versions can be a signal too. 

Often 1.0 means stability and 2.0 means shiny. For my code, a new version just means "Sorry everyone, I broke things."

> "If a complete break is required, create a new package with a new import path." - [Go FAQ][get_version]

Gustavo Niemeyer provides the Go community with a simple stateless website called [gopkg.in][]. It turns branches and tags on GitHub into new import paths.

People use a library just like any other, with `go get` and an import path.

```go
import "gopkg.in/fsnotify.v1"

// ...

watcher, err := fsnotify.NewWatcher() // note: fsnotify, not fsnotify.v1
```

When I adopted [fsnotify][] from Chris Howey, my express purpose was to *break* the API. 

Having discussed the API with members of the Go team, I had a number of changes to make. At the same time, I knew I didn't have it all figured out. The error handling needs a good long look, and integrating [new watchers](https://github.com/go-fsnotify/fsevents) will inevitably impact the API.

With version 2 and even version 3 in sight, I turned to [gopkg.in][]. It was a relatively straightforward process. Just tag each release with [semantic versions](http://semver.org/).

[fsnotify][] has some things going for it that make this simple:

* It has one dependency -- the standard library.
* The entire library is in one package. There are no absolute import paths to internal packages.

So far [gopkg.in][] has worked out quite well. One day I hope to not need it.

# Rise of the Triad

Godep and gopkg.in [aren't the only options][PackageManagementTools], but they are two community provided tools that can be used alongside `go get`. Both gopkg.in and `godep -r` allow other developers to install a project or library with `go get`. No special tools required. 

If everyone adopted Godep (or similar), there would be little need for gopkg.in. Yet they work together just fine.

It has been 1.5 years since I first wrote about [Go Package Management][history]. In that time I've learned a lot, and [some things have changed](https://github.com/golang/go). The next few years should prove interesting, but I'm really happy with what we have today.

I would like to say thank you to the people who wrote the tools that make my life easier. Thanks!

P.S. Remember *good things come in small packages.*

[code]: https://golang.org/doc/code.html
[remotes]: https://blog.splice.com/contributing-open-source-git-repositories-go/
[case-against-3pl]: /advent-2014/case-against-3pl/

[Godep]: https://github.com/tools/godep
[gopkg.in]: http://labix.org/gopkg.in

[get_version]: http://golang.org/doc/faq#get_version
[CoreOS]: https://coreos.com/blog/godep-for-end-user-go-projects/

[go-packages]: http://nathany.com/go-packages/
[fsnotify]: http://fsnotify.org/

[PackageManagementTools]: https://github.com/golang/go/wiki/PackageManagementTools

[history]: https://github.com/nathany/nathany.github.io/commits/master/_posts/2013-07-25-go-packages.md?page=2
