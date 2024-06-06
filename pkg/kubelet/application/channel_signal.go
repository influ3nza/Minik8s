package application

import "os"

var ChannelStart = make(chan string)
var ChannelEnd = make(chan os.Signal, 1)
