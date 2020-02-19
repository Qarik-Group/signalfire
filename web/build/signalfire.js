var AuthMethod;
(function (AuthMethod) {
    AuthMethod[AuthMethod["NONE"] = 0] = "NONE";
    AuthMethod[AuthMethod["USERPASS"] = 1] = "USERPASS";
})(AuthMethod || (AuthMethod = {}));
;
class APIError {
    constructor(errorMessage, returnCode) {
        this.errorMessage = errorMessage;
        this.returnCode = returnCode;
        this.error = errorMessage;
        this.code = returnCode;
    }
}
class SignalFire {
    doRequest(method, path, data) {
        return $.ajax({
            method: method,
            url: path,
            dataType: "json",
            data: (data ? JSON.stringify(data) : undefined)
        }).promise();
    }
    handleFailure(jqXHR, textStatus) {
        throw new APIError(textStatus, jqXHR.status);
    }
    authType() {
        return this.doRequest("GET", "/v1/info")
            .then(data => (data.auth_type == "userpass" ? AuthMethod.USERPASS : AuthMethod.NONE), this.handleFailure);
    }
    authNoop() {
        return this.doRequest("POST", "/v1/auth")
            .then(() => { }, this.handleFailure);
    }
    authBasic(username, password) {
        return this.doRequest("POST", "/v1/auth", {
            username: username,
            password: password
        })
            .then(() => { }, this.handleFailure);
    }
    directors() {
        return this.doRequest("GET", "/v1/directors")
            .then(data => data.directors, this.handleFailure);
    }
}
class Director {
}
$(function () {
    console.log("ready!");
    let signalfire = new SignalFire();
    doAuthThen(signalfire, showDirectors, signalfire);
});
function doAuthThen(signalfire, fn, ...args) {
    signalfire.authType()
        .then(authMethod => {
        if (authMethod == AuthMethod.NONE) {
            signalfire.authNoop()
                .then(() => fn(...args))
                .catch(() => console.log("Couldn't do no-op auth!"));
        }
        else {
            console.log("Can't do userpass right now");
        }
    })
        .catch(() => { console.log("Couldn't get auth type!"); });
}
function showDirectors(signalfire) {
    signalfire.directors()
        .then(directors => {
        let content = "<ol>";
        for (let director of directors) {
            content += "<li>" + director.name + "</li>";
        }
        content += "</ol>";
        $("#directors-list").html(content);
    })
        .catch(() => { console.log("Couldn't get director list!"); });
}
