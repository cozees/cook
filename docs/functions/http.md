# Http Functions

Http functions provide pre-define function to send get, head, options, post, patch, put and delete request to the server.

1. [get, fetch](#get-fetch)
2. [head](#head)
3. [options](#options)
4. [post](#post)
5. [put](#put)
6. [delete](#delete)
7. [patch](#patch)
## @get, @fetch

Usage:
```cook
@get [-h key:val [-h key:value] ...] URL
```

Send an http request to the server at [URL] and return a response map which contain         two key "header" and "body". The "header" is a map of response header where the "body"       is a reader object if there is data in the body. @get function can be use with redirect statement as well as assign statement.         However if the data from the function is too large it's better to use redirect       statement to store the data in a file instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -h, --header | nil | custom http header to be include or override existing header in the request. |
| --strict | false | enforce the http request and response to follow the standard of http definition for each method. |

Example:

```cook
@get -h X-Sample:123 https://www.example.com
```
[back top](#http-functions)

---

## @head

Usage:
```cook
@head [-h key:val [-h key:value] ...] URL
```

Send an http request to the server at [URL] and return a response map which contain         two key "header" and "body". The "header" is a map of response header where the "body"       is a reader object if there is data in the body. Note: By standard, head request should not have response body thus if the a restriction flag is given the        function will cause program to halt the execution otherwise a warning message is        written to standard output instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -h, --header | nil | custom http header to be include or override existing header in the request. |
| --strict | false | enforce the http request and response to follow the standard of http definition for each method. |

Example:

```cook
@head -h X-Sample:123 https://www.example.com
```
[back top](#http-functions)

---

## @options

Usage:
```cook
@options [-h key:val [-h key:value] ...] URL
```

Send an http request to the server at [URL] and return a response map which contain         two key "header" and "body". The "header" is a map of response header where the "body"       is a reader object if there is data in the body. @option function can be use with redirect statement as well as assign statement.         However if the data from the function is too large it's better to use redirect       statement to store the data in a file instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -h, --header | nil | custom http header to be include or override existing header in the request. |
| --strict | false | enforce the http request and response to follow the standard of http definition for each method. |

Example:

```cook
@options -h X-Sample:123 https://www.example.com
```
[back top](#http-functions)

---

## @post

Usage:
```cook
@post [-h key:val [-h key:value] ...] [-d data] [-f file] URL
```

Send an http request to the server at [URL] and return a response map which contain         two key "header" and "body". The "header" is a map of response header where the "body"       is a reader object if there is data in the body. @post function can be use with redirect statement as well as assign statement.         However if the data from the function is too large it's better to use redirect       statement to store the data in a file instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -h, --header | nil | custom http header to be include or override existing header in the request. |
| -d, --data | "" | string data to be sent to the server. Although, by default the data is an empty string, function will not send       empty string to the server unless it was explicit in argument with --data "". |
| -f, --file | "" | a path to a file which it's content is being used as the data to send to the server.        Note: if both flag "file" and "data" is given at the same time then flag "file" is used instead of "data". |
| --strict | false | enforce the http request and response to follow the standard of http definition for each method. |

Example:

```cook
@post -h Content-Type:application/json -d '{"key":123}' https://www.example.com
```
[back top](#http-functions)

---

## @put

Usage:
```cook
@put [-h key:val [-h key:value] ...] [-d data] [-f file] URL
```

Send an http request to the server at [URL] and return a response map which contain         two key "header" and "body". The "header" is a map of response header where the "body"       is a reader object if there is data in the body. Note: By standard, put request should not have response body thus if the a restriction flag is given the        function will cause program to halt the execution otherwise a warning message is        written to standard output instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -h, --header | nil | custom http header to be include or override existing header in the request. |
| -d, --data | "" | string data to be sent to the server. Although, by default the data is an empty string, function will not send       empty string to the server unless it was explicit in argument with --data "". |
| -f, --file | "" | a path to a file which it's content is being used as the data to send to the server.        Note: if both flag "file" and "data" is given at the same time then flag "file" is used instead of "data". |
| --strict | false | enforce the http request and response to follow the standard of http definition for each method. |

Example:

```cook
@put -h Content-Type:application/json -d '{"key":123}' https://www.example.com
```
[back top](#http-functions)

---

## @delete

Usage:
```cook
@delete [-h key:val [-h key:value] ...] [-d data] [-f file] URL
```

Send an http request to the server at [URL] and return a response map which contain         two key "header" and "body". The "header" is a map of response header where the "body"       is a reader object if there is data in the body. @delete function can be use with redirect statement as well as assign statement.         However if the data from the function is too large it's better to use redirect       statement to store the data in a file instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -h, --header | nil | custom http header to be include or override existing header in the request. |
| -d, --data | "" | string data to be sent to the server. Although, by default the data is an empty string, function will not send       empty string to the server unless it was explicit in argument with --data "". |
| -f, --file | "" | a path to a file which it's content is being used as the data to send to the server.        Note: if both flag "file" and "data" is given at the same time then flag "file" is used instead of "data". |
| --strict | false | enforce the http request and response to follow the standard of http definition for each method. |

Example:

```cook
@delete -h Content-Type:application/json -d '{"key":123}' https://www.example.com
```
[back top](#http-functions)

---

## @patch

Usage:
```cook
@patch [-h key:val [-h key:value] ...] [-d data] [-f file] URL
```

Send an http request to the server at [URL] and return a response map which contain         two key "header" and "body". The "header" is a map of response header where the "body"       is a reader object if there is data in the body. @patch function can be use with redirect statement as well as assign statement.         However if the data from the function is too large it's better to use redirect       statement to store the data in a file instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -h, --header | nil | custom http header to be include or override existing header in the request. |
| -d, --data | "" | string data to be sent to the server. Although, by default the data is an empty string, function will not send       empty string to the server unless it was explicit in argument with --data "". |
| -f, --file | "" | a path to a file which it's content is being used as the data to send to the server.        Note: if both flag "file" and "data" is given at the same time then flag "file" is used instead of "data". |
| --strict | false | enforce the http request and response to follow the standard of http definition for each method. |

Example:

```cook
@patch -h Content-Type:application/json -d '{"key":123}' https://www.example.com
```
[back top](#http-functions)

---

