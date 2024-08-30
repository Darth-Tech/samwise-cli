## samwise checkForUpdates ci

For CI integrations[experimental]

### Synopsis


	
	Includes features for better CI integrations such as failure when updates available 
	for pipelines, allowing users to automatically create PRs when updates are present(custom thresholds) and so on.

Not all those who don't update dependencies are lost.

```
samwise checkForUpdates ci [flags]
```

### Options

```
  -h, --help   help for ci
```

### Options inherited from parent commands

```
      --config string            config file (default is $HOME/.samwise.yaml)
  -d, --depth int                Folder depth to search for modules in. Give -1 for a full directory extraction.
      --git-repo string          Git Repository to check module dependencies on. (default "g")
  -i, --ignore stringArray       Directories to ignore when searching for the One Ring(modules and their sources. (default [.git,.idea])
  -o, --output string            Output format. Supports "csv" and "json". Default value is csv. (default "csv")
  -f, --output-filename string   Output file name. (default "module_report")
      --path string              The path for directory containing terraform code to extract modules from. (default "p")
  -v, --verbose                  The path for directory containing terraform code to extract modules from.
```

### SEE ALSO

* [samwise checkForUpdates](samwise_checkForUpdates.md)	 - search for updates for terraform modules using in your code and generate a report

