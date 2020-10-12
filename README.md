# fget
Multithread download for a list of files.

### install
```
go get -u -v github.com/bp0lr/fget
```

### examples
```
┌─[root@DarkStar]─[/opt/bp0/lovan/fget]
└──╼ #cat all_jslinks.txt | fget -o /opt/bp0/lovan/fget/ -H "foo: bar" --proxy "http://192.168.0.150:8080"
```

```
┌─[root@DarkStar]─[/opt/bp0/lovan/fget]
└──╼ #cat all_jslinks.txt | fget -o /tmp/myresults --no-folders --random-agent
```

### usage
```
┌─[✗]─[root@DarkStar]─[/opt/bp0/lovan/fget]
└──╼ #go run main.go -h
Usage of /tmp/go-build547111181/b001/exe/main:
  -f, --follow-redirect      Follow redirects (Default: false)
  -H, --header stringArray   Add custom Headers to the request
  -o, --output string        Target directory (default "Save to folder. Default: create results folder which will include a folder for each target")
      --no-folders           Don't store results on separate folders
  -p, --proxy string         Add a HTTP proxy
  -r, --random-agent         Set a random User Agent
  -u, --url string           The url to check
  -v, --verbose              Display extra info about what is going on
  -w, --workers int          Number of workers (default 20)
```
