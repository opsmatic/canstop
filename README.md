**WARNING**: This is very much a work in progress. I highly advise you do not
import this project yet, as I the API likely _WILL_ change in the near future.

### Can't Stop Won't Stop.. Or Can You?

This is a fairly simple package for managing the lifecycle of daemon processes
within a Go program. It builds a bit on top of the [`tomb`](http://launchpad.net/tomb) package.

It is inspired by:

* My work at UA with Neil Walker on making services clean up after themselves
* Dropwizard's [`lifecycle`](https://github.com/codahale/dropwizard/tree/master/dropwizard-lifecycle/src/main/java/com/codahale/dropwizard/lifecycle)
* The need to do graceful shutdown/cleanup across multiple Go projects
* Lots of long talks with [Richard Crowley](http://rcrowley.org/articles/golang-graceful-stop.html)

### Usage

Nuking this section for now because too much is changing. See `example` and
`example_echo` directories, as well as the tests.
