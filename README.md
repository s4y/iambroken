[iambroken.com](https://iambroken.com/) and [istheinternetdown.com](https://istheinternetdown.com/) are small utility websites I run.

[istheinternetdown.com](https://istheinternetdown.com/) serves a webpage with a random background color to browsers, and plain text otherwise. Both include UTC time and the client's IP address.


[iambroken.com](https://iambroken.com/) has a number of subdomains which each return one piece of information:

| Subdomin | What it returns |
| --- | --- |
| [ip.iambroken.com](https://ip.iambroken.com) | Your IP address |
| [ua.iambroken.com](https://ua.iambroken.com) | Your user agent |
| [time.iambroken.com](https://time.iambroken.com) | UTC time (for teasing out unexpected caching) |
| [echo.iambroken.com](https://echo.iambroken.com) | Your HTTP request (currently not working as well a I’d like) |
| `loop.iambroken.com` | Resolves to `127.0.0.1` (not represented in this repo) |

All responses are plain text with a trailing newline, for easy `curl`ing.


Pull requests and feature requests welcome.
