Testing this locally is a huge pain in the ass so here is what I've gotten set up to make this (mostly) work.

1 - CORS
Use the cors-anywhere project in conjunction with heroku to run a local copy of the app.

link to repo - https://github.com/Rob--W/cors-anywhere

Steps to use:
  1 - Copy the repo down either by just downloading the zip from the above repo, or by doing a `git clone` on the library
  2 - Make sure you have heroku installed. If not here's a link to the getting started docs (https://devcenter.heroku.com/articles/getting-started-with-nodejs#set-up)
  3 - Once heroku is installed open the folder the repo is in in your CLI of choice and run `heroku local` and that's it. You'll see the server running on port 5000 by default accepting connections from anywhere. Since this is local testing we don't care about locking it down but there are more config options that can be provided if desired.

2 - Static content
This site heavily relies on the static content side. The backend websocket messaging is just ferrying messages and providing some light security at this point. You'll want to set up another server to serve you static content now.

Steps to set up:
  1 - Make sure you have Python 3 installed and set up
  2 - Open the directory that you have the https://github.com/mpuckett159/stack-web-app-static saved into
  3 - Run `python -m http.server 8000` Note that this can be any port at the end I just use 8000 personally
  4 - To make use of the cors-anywhere server you will also need to update the code in the app.js file to append `http://localhost:5000/` on the front of everything, e.g. the following line would be updated as follows:
  ```
  this.meetingUrl = window.location.host + '/?meeting_id=' + this.tableId;
  becomes
  this.meetingUrl = 'http://localhost:5000/' + window.location.host + '/?meeting_id=' + this.tableId;
  ```
  This is very cumbersome and difficult and introduces a lot of potential issues when finalizing and pushing changes to GitHub. I have no idea how to get around those pit falls at this point and am open to suggestions if anyone knows.

3 - Go backend server
Thank God this is easy

link to repo - https://github.com/mpuckett159/stack-web-app

Steps to set up:
  1 - Open the directory the repo is saved in
  2 - Build the container if you haven't already `docker build -t <tag-name> .`
  3 - Set your environment variables (user a .env file for it):
    * PORT = 8080
    * DEBUG = true
    * DISABLEWEBSOCKETORIGINCHECK = true
  The last one is unfortunately required to test locally. cors-anywhere doesn't allow websocket origin checking validation as far as I know.
  4 - Finally run `docker run --env-file .env -p 8080:8080 <tag-name>` and there you have it. 3 running servers to test the application.