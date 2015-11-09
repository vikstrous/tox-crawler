# Tox Crawler

This project is a crawler of the tox network. It's built in Go using the [gotox](https://github.com/vikstrous/gotox) library.

See it live at http://tox.viktorstanchev.com.

To deploy:

```
docker run --name tox-crawler -d -p 80:7071 --restart always -v $(pwd)/data:/go/src/github.com/vikstrous/tox-crawler/data vikstrous/tox-crawler
```

Example of a graphs it can produce right now:

![screenshot](https://raw.githubusercontent.com/vikstrous/tox-crawler/master/screenshot.png)

![screenshot](https://raw.githubusercontent.com/vikstrous/tox-crawler/master/screenshot2.png)
