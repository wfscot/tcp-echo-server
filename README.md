# tcp-echo-server
Simple TCP server that accepts unlimited connections and echos any data received back to the client immediately.

Additional support available for periodically sending "alive" to the client regardless of whether data is received.

Includes both dedicated server and echoer objects that can be used directly as well as a lightweight CLI application.

## CLI Application

### Building

Standard go build process.  Download or clone the repo and from the main directory `go build`

### Usage

```
Usage: tcp-echo-server [-aqv] listenPort
 -a, --announcealive
       announce "alive" every 5 seconds.
 -q    quiet. do not print any log info. overrides verbosity flag.
 -v    verbosity. can be used multiple times to further increase.
 ```
 
 The only required argument is the port on which to listen.  Note that the applicaiton will always listen on all interfaces.  Specific interface support is not (yet) implemented.
 
 ## Reusing Objects Directly
 
 It's alo entirely possible to instantiate either a Server or Echoer in another application by using the objects with the `echoer` package.
 
 In your code, import `github.com/wfscot/tcp-echo-server/echoer`.  You don't need anything from the main directory or package.
 
 Note that both Echoer and Server require a Context to run. Cancelling that Context will cleanly tear down everything.  Furthermore, logging is implemented via a zerolog Logger instance stored in the Context via the zerolog standard Logger.WithContext() mechanism.  If the Logger is not found, logging will be disabled.  Please look to `main.go` for an example of how to do this.
