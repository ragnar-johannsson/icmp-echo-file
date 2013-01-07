## ICMP Echo File

A utility that echos the contents of a file inside the data segment in a stream of ICMP echo requests. It's left as an exercise for the reader to judge the usefulness of this endeavour.


### Installation & usage

```
$ go get github.com/ragnar-johannsson/icmp-echo-file
```

This will install the binary `icmp-echo-file` in you $GOPATH/bin. Then, simply:

```
$ sudo icmp-echo-file example.org file.txt
```

See `icmp-echo-file -h` for more info on usage.


### License

Simplified BSD. See the LICENSE file for further information.
