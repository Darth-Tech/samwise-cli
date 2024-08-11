## samwise checkForUpdates

search for updates for terraform modules using in your code and generate a report

### Synopsis

Searches (sub)directories for module sources and versions to create a report listing versions available for updates.

CSV format : repo_link | current_version | updates_available

JSON format: [{
                "repo_link": <repo_link>,
                "current_version": <current version used in the code>,
                "updates_available"
             }]
	

```
samwise checkForUpdates [flags]
```

### Options

```
  -d, --depth int                Folder depth to search for modules in. Give -1 for a full directory extraction.
      --git-repo string          Git Repository to check module dependencies on. (default "g")
  -h, --help                     help for checkForUpdates
  -i, --ignore stringArray       Directories to ignore when searching for the One Ring(modules and their sources. (default [.git,.idea])
  -o, --output string            Output format. Supports "csv" and "json". Default value is csv. (default "csv")
  -f, --output-filename string   Output file name. (default "module_report")
      --path string              The path for directory containing terraform code to extract modules from.
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.samwise.yaml)
```

### SEE ALSO

* [samwise](samwise.md)	 - A CLI application to accompany on your terraform module journey and sharing your burden of module dependency updates, just as one brave Hobbit helped Frodo carry his :)

