# gochanscraper

* Installation
```sh
  go install github.com/Exodus76/gochanscraper@latest
```

* Usage
```
  gochanscraper https://boards.4chan.org/<board>/thread/<id>
```
 creates a folder with thread_id as folder name<br/><br/>

* Compile
```sh
  git clone https://github.com/Exodus76/gochanscraper.git
  cd gochanscraper/build
  go build ../
```

# TODO
- implement watch function to autodownload when new posts
- maybe give a notification when the thread has expired so the user knows that there are no more posts to autodownload