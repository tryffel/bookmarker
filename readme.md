# Bookmarker
Bookmarker is a terminal application to manage and view bookmarks. It is under development and not stable yet.

# Features
* Assign any key-value metadata (currently editable in config file) 
* Advanced searching. Search can be simple like 'bookmark', or more advanced: 'author:davis project:study link:archives.com'
* Store IPFS links directly with corresponding bookmark
* Import existing bookmarks (no exporting yet)
* Customize color scheme
* Archived status 
* Sort bookmarks

# Searching & filtering
If built with support for sqlite fts5 extension, bookmarker supports full text search queries. Full text queries 
currently cover bookmark name, description, project & link content. 
Some examples of full text queries that are supported:
```
'help page' -> match any phrase that has words help and page
"help page" -> match any phrase that has phrase "help page"
'help pag*' -> match any phrase that has help and pag*, where * is wildcard
'help AND page OR site'
'^help' -> phrase must start with help
```

for more info see [Sqlite FTS5 extension](https://www.sqlite.org/fts5.html)

Filtering narrows down results with simple key-value pairs:
```
author:dave language:english -link:mypage.com -> author & language must match, link cannot contain given
```

# Building
Assuming go already installed, run
```
go get -u tryffel.net/go/bookmarker
# run
GOPATH/bin/bookmarker
```

If you wish to enable full text search capability, you need to give following flag during compilation:
```
go build --tags 'fts5'
```

Config file can be set with ```--config``` flag. This will create new file and directories, if they don't exist. 
Usage:
```
./bookmarker --config /my/dir/config.toml
```

After this, all data will be at directory /my/dir by default. This can be customised in config file.
