**WARNING**: This is very much a work in progress. I highly advise you do not
import this project yet, as I the API likely _WILL_ change in the near future.

### Can't Stop Won't Stop.. Or Can You?

This is a fairly simple package for managing the lifecycle of daemon processes
within a Go program.

##### Motivation

When building long-running programs, one often deals with  multiple concurrent
processes that are responsible for different portions of the program's
functionality. For example, the thread that accepts incoming connections on a
socket, or threads that pull messages off of a queue and perform some operation
on it. When dealing with such processes, it is usually desirable to allow them
to finish their work cleanly when the program is terminated - stop accepting new
connections and finish serving the ones already in flight; stop pulling jobs off
the queue, but finish the ones that are already in progress; etc.

The general pattern is to spin up a bunch of background processes/threads/goroutines in
Go's case in the `main()` function and  wait for some event that indicates it's time
to stop (usually SIGINT or similar). Once this event occurs, the background
tasks are notified of this event (in Java's case, a good example is
`ExecutorService.shutdown`) and the `main()` thread waits for some time for them
to clean up (`ExecutorService.awaitTermination`). Any stragglers are then
abandoned, hopefully with a nasty log message.

Richard Crowley and I (and probably some other folks) spent some time talking
about how this might be done in Go, and he eventually produced [this blog
post](http://rcrowley.org/articles/golang-graceful-stop.html) detailing a way to
use channels to signal shutdown to background processes. The main trick is that
a call to `close()` causes calls from any number of goroutines on that channel
to return a value. If the only thing that you ever do on this channel is
`close()` it, you can rely on that return value to indicate that the channel's
been closed. This event can then be the signal that it's time to close up shop.

I spent a long time thinking about how to make this generalizable, and it took
me a while to actually trace my way back to "the requirements." Here's what I
came up with along with some terms that help me think about it:

* **Services**: Certain processes should run for the duration of the program, and should be
signaled to shut down when it's time for the program to quit.
* **Sessions**: Other processes are started and stopped many times naturally during the course of
the program (think goroutines that handle an individual connection), and don't
need to be tied to the program's lifetime; however, they should be given time to
clean up before `main()` exits, leaving them in an undetermined state (think a
request in progress)
* When Services terminate un-cleanly, we should complain noisily, as that
likely something unpleasant happened with 1 or more Session (a corrupted
response, etc)

##### Inspiration:

* My work at UA with Neil Walker on making services clean up after themselves
* Dropwizard's [`lifecycle`](https://github.com/codahale/dropwizard/tree/master/dropwizard-lifecycle/src/main/java/com/codahale/dropwizard/lifecycle)
* The need to do graceful shutdown/cleanup across multiple Go projects
* Lots of long talks with [Richard Crowley](http://rcrowley.org/articles/golang-graceful-stop.html)

#### Implementation

*(CURRENTLY IN FLUX)*

### Usage

Nuking this section for now because too much is changing. See `example` and
`example_echo` directories, as well as the tests.
