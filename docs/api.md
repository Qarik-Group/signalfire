## GET /v1/info

```json
{
  "version": "0.0.1",
  "auth_type": "basic"
}
```

## GET /v1/deployment-groups

### Response:

```json
{
  "groups": [
    {
      "name": "bosh",
      "deployments" : [
        {
          "id": "c28e361e3272f8534835b50dff93d7b4c8c87956",
          "name": "snw-dev-bosh",
          "director_id": "01234567-89ab-cdef-0123-456789abcdef",
        },
        {
          "id": "2b80b79b41cc60e44d6a53231288bb6f61dedb5b",
          "name": "snw-prod-bosh",
          "director_id": "89abcdef-0123-4567-89ab-cdef01234567",
        }
      ],
      "releases": [
        {
          "name": "bosh",
          "versions": [
            {
              "version": "270.2",
              "deployments": [
                "2b80b79b41cc60e44d6a53231288bb6f61dedb5b"
              ]
            },
            {
              "version": "270.9",
              "deployments": [
                "c28e361e3272f8534835b50dff93d7b4c8c87956"
              ]
            }
          ]
        },
        {
          "name": "uaa",
          "versions": [
            {
              "version": "72.0.0",
              "deployments": [
                "c28e361e3272f8534835b50dff93d7b4c8c87956",
                "2b80b79b41cc60e44d6a53231288bb6f61dedb5b"
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### GET /v1/directors

## Response

```json
{
  "directors": [
    {
      "name": "snw-proto-bosh",
      "uuid": "01234567-89ab-cdef-0123-456789abcde",
    },
    {
      "name": "snw-dev-bosh",
      "uuid": "89abcdef-0123-4567-89ab-cdef01234567",
    }
  ]
}
```
