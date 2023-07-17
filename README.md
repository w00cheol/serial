# Serial communication with DLP-TH1C
[GO PACKAGE LINK](https://pkg.go.dev/github.com/w00cheol/serial)

Provides only functions that communicate using ascii.  
In my environment, using byte, the dlp-th1c sensor loses some data for some reason (but I couldn't find).  
The "//string parsing" parts in various parts of the function were also written considering data loss.  

Select function by the command needs to be updated to use onther parsing function depending on each option.  
But imagine how the custom option could be,  
The value of the sensor responds is too variable to expect every kind of format, exception, data loss as well.  
So it is decided to call readAllAsync(chan) function and just extract only the kind of data that user wants.  


### USAGE
```console
    go mod init (YOUR_MODULE_NAME)
    go get github.com/w00cheol/serial
```  

```go
package main

import "github.com/w00cheol/serial"

func main() {
    serial.InitPort("PORTNAME_AS_STRING") // e.g) "/dev/ttyACM1", default is "/dev/ttyACM0"
    serial.RunWithCommand("COMMAND_AS_STRING") // e.g) "t"
}
```  

```console
    go mod tidy
    go run .
```


|COMMAND        |FUNCTION                                   |
|--------------:|:------------------------------------------|
| all           | Read All Data                             |      
|(COMBINE)      | Read Costomized Data but takes 30 secs    |
|t              | Read Temperature Data Only                |
|h              | Read Humidity Data Only                   |
|p              | Read Pressure Data Only                   |
|a              | Read Tilt Data Only                       |
|x              | Read Vibration (X Axis) Data Only         |
|v              | Read Vibration (Y Axis) Data Only         |
|w              | Read Vibration (Z Axis) Data Only         |
|l              | Read Light Level Data Only                |
|f              | Read Sound Data Only                      |
|b              | Read Broadband Data Only                  |

LICENSE: Apache-2.0 