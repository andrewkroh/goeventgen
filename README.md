# goeventgen

Writes events to a Windows event log using data from a text file. It creates a
new event record for each line of text that it reads from the specified file.

### Usage

```
PS C:\Users\vagrang> goeventgen-amd64.exe -h
Usage of goeventgen.exe:
  -f string
        file to read
  -id int
        event id (default 512)
  -l string
        event source name (default "EventSystem")
  -max uint
        maximum events to write
```

### Download

[Download](https://github.com/andrewkroh/goeventgen/releases/) binaries from the
Github releases page.

### Example Source Data

You can use the [HTTP logs from
NASA](http://ita.ee.lbl.gov/html/contrib/NASA-HTTP.html).
