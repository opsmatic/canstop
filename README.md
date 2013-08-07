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

Create a `Runner` using `canstop.NewRunner`, then wrap your work in the
following boilerplate:

```go
type YourThingToManage struct {}

func (self *YourThingToManage) Run(t *tomb.Tomb) {
	for {
		select {
		case _ = <- t.Dying():
			{
				// do your cleanup, call t.Done() if you're feeling responsible
			}
		default: // leave this empty and fall through to actually doing work
		}

		// Do your work here.. read from another channel, whatever
		// if you encounter a terrible fatal error, call t.Kill()
	}
}
```

Now just pass that thing to your `Runner` and watch it **GOOOOOO**

```go
thing := &YourThingToManage{}
r.RunMe(thing)
// probably some code to catch singals to know when to call r.Stop()
r.Wait()
```

### Example

There's an example in the `example` dir. It has 2 long running processes, a
Worker and a Producer. The Worker is set up correctly, the Producer is not and
ends up getting bludgeoned after a timeout. `canstop` then reports this fact to
you in the logs.

```
$ go run example/example.go
2013/08/07 04:47:40 Found another multiple of 4! 5577006791947779410
2013/08/07 04:47:41 Found another multiple of 4! 8674665223082153551
2013/08/07 04:47:42 Found another multiple of 4! 6129484611666145821
^C2013/08/07 04:47:43 Clean shut down of worker. Found 3 matches
2013/08/07 04:47:47 Ungraceful stop: Job took too long to terminate, forcing termination after 5000000000
```

Note that the cleanup code was called correctly and that the process took 5
seconds to terminate because the producer just kept chugging along (5 seconds is
the timeout we set `Runner` with)

### Next Steps

I think the amount of boilerplate required could be vastly reduced by embedding
`tomb.Tomb` in a wrapper that provides some gracefulness-specific functionality.
Might be possible to provide some convenience methods where all the caller
actually has to provide is the inner body of the work for loop, somehow
indicating to them that every call to their function will be done inside a tight
loop that will also be checking for cancellation at every iteration.

I also think this project needs to provide an optional panic protection wrapper, thus
making it possible to log when things freak out and break, but be able to
restart them.

Accounting is also somewhat weak - perhaps there should be a way to extract the
error values from the `Runner`. There's also no way to know which job failed or
was terminated prematurely the way things are currently set up. If I wrap
`tomb.Tomb` in the aforementioned `canstop` specific struct, I can pass the
`.String()` of the argument to `RunMe` into it, making it possible for the
particulaly dilligent user to get readable debug logs.
