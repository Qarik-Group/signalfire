/*==============================
  A HTTP Client for SignalFire
==============================*/

/**
 * Enumeration of supported auth methods
 */
enum AuthMethod { NONE, USERPASS };

/**
 * Contains error information returned from an HTTP API call
 */
class APIError {
  readonly error: string;
  readonly code: number;

  constructor(readonly errorMessage: string, readonly returnCode: number) {
    this.error = errorMessage;
    this.code = returnCode;
  }
}

/**
 * An HTTP Client to the SignalFire API
 */
class SignalFire {
  private doRequest(method: string, path: string, data?: object): JQuery.Promise<any> {
    return $.ajax({
      method: method,
      url: path,
      dataType: "json",
      data: (data ? JSON.stringify(data) : undefined)
    }).promise();
  }

  private handleFailure(jqXHR: JQuery.jqXHR<any>, textStatus: string): never {
    throw new APIError(textStatus, jqXHR.status);
  }

  authType(): Promise<AuthMethod> {
    return this.doRequest("GET", "/v1/info")
      .then(
        data => (data.auth_type == "userpass" ? AuthMethod.USERPASS : AuthMethod.NONE),
        this.handleFailure
      );
  }

  authNoop(): Promise<void> {
    return this.doRequest("POST", "/v1/auth")
      .then(
        () => { },
        this.handleFailure
      )
  }

  authBasic(username: string, password: string): Promise<void> {
    return this.doRequest("POST", "/v1/auth", {
      username: username,
      password: password
    })
      .then(
        () => { },
        this.handleFailure
      );
  }

  directors(): Promise<Array<Director>> {
    return this.doRequest("GET", "/v1/directors")
      .then(
        data => data.directors as Array<Director>,
        this.handleFailure);
  }
}

class Director {
  name: string;
  uuid: string;
}