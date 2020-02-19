/// <reference path="./signalfire.ts"/>


$(function () {
  console.log("ready!")

  let signalfire = new SignalFire();
  doAuthThen(signalfire, showDirectors, signalfire)

})

function doAuthThen(signalfire: SignalFire, fn: Function, ...args: any[]) {
  signalfire.authType()
    .then(authMethod => {
      if (authMethod == AuthMethod.NONE) {
        signalfire.authNoop()
          .then(() => fn(...args))
          .catch(() => console.log("Couldn't do no-op auth!"));
      } else {
        console.log("Can't do userpass right now")
      }
    })
    .catch(() => { console.log("Couldn't get auth type!"); });
}

function showDirectors(signalfire: SignalFire) {
  signalfire.directors()
    .then(directors => {
      let content: string = "<ol>";
      for (let director of directors) {
        content += "<li>" + director.name + "</li>";
      }
      content += "</ol>";
      $("#directors-list").html(content);
    })
    .catch(() => { console.log("Couldn't get director list!"); });
}