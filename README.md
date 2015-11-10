# Stargraph
## The perfect tool to plot graph of stars on Github repositories

There is a CLI where you provide your Github API token via `-t` and the `:username/:reponame` via the `-r` parameters.
If you don't have a Github API token, you will be restricted by the API to 60 calls per hours.

An example on how to use it:
```
stargraph -t githubtoken -r evermax/stargraph
```

To get the project, just do `go get github.com/evermax/stargraph`

The program will produce 3 files:

 - graph.png which contains the graph plotted with [gonum/plot](https://github.com/gonum/plot). It currently doesn't support date display so you will end up with Unix timestamp on the X-axis
 - canvasDB.json which contains graph data to be used with [CanvasJS](http://canvasjs.com)
 - jqplotDB.json which contains graph data to be used with [jqplot](http://www.jqplot.com)

You can also now use the lib part of the project to get the timestamps of the stars on a repository as a `[]int64`

## Disclaimer
This tool only take the current stars on the repository and place them on a graph
where their are placed by order of apparences. That is why it will never provide you with a shrinking graph.

It is still a funny way to see it the repo has a good growth. You just need to pay attention to the last star timestamp.

## Demo
I made a proof of concept for this project on [Stargraph.co](http://stargraph.co). You can there provide a Github API token or log in with Github and try it!

## BaSiProMo
I want to make this project better during [GaSiProMo](https://codelympics.io/projects/3) by:

 - [ ] clean up the code
 - [ ] write documentation and put it on godoc
 - [x] Make the project a library using TDD
   - [x] Seperate the image, the Canvas JSON file and the jqplot JSON file
   - [x] Provide a writer to the library to write the image to
   - [ ] Find a plotting library that can have a time scale
 - [x] Make a presentation website about the project
 - [ ] Make a web plateform to which you can connect via Github and graph your project
   - [ ] Have a way on that plateform to provide a image of this graph that can be used on the README of the repo
   - [ ] Automatically and frequently rebuild the graph for the ones that were already crawled


## Contributions
Contributions to that project are more than welcome, especially... for everything!
You are welcome to talk to me about it on [Golang slack](https://gophers.slack.com/messages/@maxime/). (If there is a need for it, I might make a channel)
[Request an invite](http://bit.ly/go-slack-signup) if you are not already on the channel.
