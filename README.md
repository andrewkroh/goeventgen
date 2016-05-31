# goeventgen

Writes events to a Windows event log using data from a text file. It creates a
new event record for each line of text that it reads from the specified file.

### Usage

```
PS C:\Users\vagrang> goeventgen-amd64.exe -h
Usage of C:\Users\Administrator\Downloads\goeventgen-amd64.exe:
  -f string
        file to read
  -id uint
        event id (default 512)
  -install
        install new event source
  -l string
        event source name (default "EventSystem")
  -max uint
        maximum events to write
  -provider string
        provider name to install (default "Application")
  -rate int
        rate limit in events per second
  -source string
        source name to install, must be specified to install new event source
```

### Download

[Download](https://github.com/andrewkroh/goeventgen/releases/) binaries from the
Github releases page.

### Example Source Data

You can use the [HTTP logs from
NASA](http://ita.ee.lbl.gov/html/contrib/NASA-HTTP.html).
