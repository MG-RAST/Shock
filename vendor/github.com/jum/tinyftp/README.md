tinyftp.go
==========

A small FTP package [go](http://golang.org) language, it uses the
net/textproto package for most of the communication.

Currently only passive operation is supported, and the connection
management is mostly outside of the package to allow using it with the
appengine/socket package.
