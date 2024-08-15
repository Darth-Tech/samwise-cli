# samwise

A CLI application to accompany on your terraform module journey and sharing your burden of module dependency updates, just as one brave Hobbit helped Frodo carry his :)


[![Go Test](https://github.com/thundersparkf/samwise-cli/actions/workflows/go-test-workflow.yml/badge.svg)](https://github.com/thundersparkf/samwise-cli/actions/workflows/go-test-workflow.yml)

```
                       \ : /
                    '-: __ :-'
                    -:  )(_ :--
                    -' |r-_i'-
            ,sSSSSs,   (2-,7
            sS';:'`Ss   )-j
           ;K e (e s7  /  (
            S, ''  SJ (  ;/
            sL_~~_;(S_)  _7
|,          'J)_.-' />'-' `Z
j J         /-;-A'-'|'--'-j\
 L L        )  |/   :    /  \
  \ \       | | |    '._.'|  L
   \ \      | | |       | \  J
    \ \    _/ | |       |  ',|
     \ L.,' | | |       |   |/
    _;-r-<_.| \=\    __.;  _/
      {_}"  L-'  '--'   / /|
            |   ,      |  \|
            |   |      |   ")
            L   ;|     |   /|
           /|    ;     |  / |
          | |    ;     |  ) |
         |  |    ;|    | /  |
         | ;|    ||    | |  |
         L-'|____||    )/   |
             % %/ '-,- /    /
             |% |   \%/_    |
          ___%  (   )% |'-; |
        C;.---..'   >%,(   "'
                   /%% /
                  Cccc'

(cause I couldn't find good enough Samwise Gamgee ASCII art)
                                        Frodo by Shanaka Dias
```

## Challenge
As your repositories grow and you reference your modules in other repositories, you would reasonably version your modules to ensure that upstream changes in the source doesn't break your infrastructure. However, it is difficult to keep track of all the new releases for the modules being used and even harder to do it regularly. Unaddressed, this builds overtime as tech debt as one day you discover that a core module is now 3 major versions ahead.

## Solution
`samwise` Searches your repository for usages of modules and generates a report of the modules that have updates available along with all the versions that are more advanced than the version used currently.
## Install instructions
### Homebrew
```
brew tap darth-tech/tap
brew install samwise-cli
```

### From source
```shell
git clone https://github.com/Darth-Tech/samwise-cli
cd samwise-cli
cp .samwise.yaml.example
```
Update the required git user token and then build using:
```shell
go build
chmod +x samwise-cli
```
This can then be moved to any of the directories in the PATH variable so that it can be used easily or it can be used in the same directory itself.
## Usage

Available Commands:
```
Available Commands:
  checkForUpdates search for updates for terraform modules using in your code and generate a report
  completion      Generate the autocompletion script for the specified shell
  help            Help about any command

Flags:
      --config string   config file (default is $HOME/.samwise.yaml)
  -h, --help            help for samwise
  -t, --toggle          Help message for toggle
```

For more details, checkout the ```docs``` folder or click [here](https://github.com/thundersparkf/samwise-cli/blob/main/docs/samwise.md)


