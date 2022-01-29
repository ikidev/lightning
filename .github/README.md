


# This is a custom version of Fiber. Still in pre-alpha stage. 


## Consider donating to [Fiber](https://buymeacoff.ee/fenny)






## ⚡️ Quickstart

```
Placeholder
```

## 🤖 Benchmarks

```
Placeholder
```

## ⚙️ Installation

```
Placeholder
```

## 🎯 Features

```
Placeholder
```

## 💡 Philosophy

```
Placeholder
```

## ⚠️ Limitations
```
Placeholder
```
## 👀 Examples

```
Placeholder
```
#### 📖 [**Basic Routing**](https://docs.gofiber.io/#basic-routing)

```
Placeholder
```

#### 📖 [**Route Naming**](https://docs.gofiber.io/api/app#name)

```
Placeholder
```

#### 📖 [**Serving Static Files**](https://docs.gofiber.io/api/app#static)

```
Placeholder
```

#### 📖 [**Middleware & Next**](https://docs.gofiber.io/api/ctx#next)

```
Placeholder
```


### Views engines

```
Placeholder
```

### Grouping routes into chains

📖 [Group](https://docs.gofiber.io/api/app#group)

```
Placeholder
```

### Middleware logger

```
Placeholder
```

### Cross-Origin Resource Sharing (CORS)

```
Placeholder
```

### Custom 404 response
```
Placeholder
```

### JSON Response
```
Placeholder
```

### WebSocket Upgrade

```
Placeholder
```

### Recover middleware
```
Placeholder
```

### Using Trusted Proxy

📖 [Config](https://docs.gofiber.io/api/fiber#config)

```go
import (
    "github.com/ikidev/lightning"
    "github.com/ikidev/lightning/middleware/recover"
)

func main() {
    app := fiber.New(fiber.Config{
        EnableTrustedProxyCheck: true,
        TrustedProxies: []string{"0.0.0.0", "1.1.1.1/30"}, // IP address or IP address range
        ProxyHeader: fiber.HeaderXForwardedFor},
    })

    // ...

    log.Fatal(app.Listen(":3000"))
}
```

## 🧬 Internal Middleware

Here is a list of middleware that are included within the Fiber framework.

| Middleware                                                                       | Description                                                                                                                                                                  |
| :------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [basicauth](https://github.com/gofiber/fiber/tree/master/middleware/basicauth)   | Basic auth middleware provides an HTTP basic authentication. It calls the next handler for valid credentials and 401 Unauthorized for missing or invalid credentials.        |
| [compress](https://github.com/gofiber/fiber/tree/master/middleware/compress)     | Compression middleware for Fiber, it supports `deflate`, `gzip` and `brotli` by default.                                                                                     |
| [cache](https://github.com/gofiber/fiber/tree/master/middleware/cache)           | Intercept and cache responses                                                                                                                                                |
| [cors](https://github.com/gofiber/fiber/tree/master/middleware/cors)             | Enable cross-origin resource sharing \(CORS\) with various options.                                                                                                          |
| [csrf](https://github.com/gofiber/fiber/tree/master/middleware/csrf)             | Protect from CSRF exploits.                                                                                                                                                  |
| [filesystem](https://github.com/gofiber/fiber/tree/master/middleware/filesystem) | FileSystem middleware for Fiber, special thanks and credits to Alireza Salary                                                                                                |
| [favicon](https://github.com/gofiber/fiber/tree/master/middleware/favicon)       | Ignore favicon from logs or serve from memory if a file path is provided.                                                                                                    |
| [limiter](https://github.com/gofiber/fiber/tree/master/middleware/limiter)       | Rate-limiting middleware for Fiber. Use to limit repeated requests to public APIs and/or endpoints such as password reset.                                                   |
| [logger](https://github.com/gofiber/fiber/tree/master/middleware/logger)         | HTTP request/response logger.                                                                                                                                                |
| [pprof](https://github.com/gofiber/fiber/tree/master/middleware/pprof)           | Special thanks to Matthew Lee \(@mthli\)                                                                                                                                     |
| [proxy](https://github.com/gofiber/fiber/tree/master/middleware/proxy)           | Allows you to proxy requests to a multiple servers                                                                                                                           |
| [requestid](https://github.com/gofiber/fiber/tree/master/middleware/requestid)   | Adds a requestid to every request.                                                                                                                                           |
| [recover](https://github.com/gofiber/fiber/tree/master/middleware/recover)       | Recover middleware recovers from panics anywhere in the stack chain and handles the control to the centralized[ ErrorHandler](https://docs.gofiber.io/guide/error-handling). |
| [timeout](https://github.com/gofiber/fiber/tree/master/middleware/timeout)       | Adds a max time for a request and forwards to ErrorHandler if it is exceeded.                                                                                                |

## 🧬 External Middleware

List of externally hosted middleware modules and maintained by the [Fiber team](https://github.com/orgs/gofiber/people).

| Middleware                                        | Description                                                                                                                                                         |
| :------------------------------------------------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [adaptor](https://github.com/gofiber/adaptor)     | Converter for net/http handlers to/from Fiber request handlers, special thanks to @arsmn!                                                                           |
| [helmet](https://github.com/gofiber/helmet)       | Helps secure your apps by setting various HTTP headers.                                                                                                             |
| [jwt](https://github.com/gofiber/jwt)             | JWT returns a JSON Web Token \(JWT\) auth middleware.                                                                                                               |
| [keyauth](https://github.com/gofiber/keyauth)     | Key auth middleware provides a key based authentication.                                                                                                            |
| [rewrite](https://github.com/gofiber/rewrite)     | Rewrite middleware rewrites the URL path based on provided rules. It can be helpful for backward compatibility or just creating cleaner and more descriptive links. |
| [session](https://github.com/gofiber/session)     | This session middleware is build on top of fasthttp/session by @savsgio MIT. Special thanks to @thomasvvugt for helping with this middleware.                       |
| [template](https://github.com/gofiber/template)   | This package contains 8 template engines that can be used with Fiber `v1.10.x` Go version 1.13 or higher is required.                                               |
| [websocket](https://github.com/gofiber/websocket) | Based on Fasthttp WebSocket for Fiber with Locals support!                                                                                                          |

## 🌱 Third Party Middlewares

This is a list of middlewares that are created by the Fiber community, please create a PR if you want to see yours!

-   [arsmn/fiber-swagger](https://github.com/arsmn/fiber-swagger)
-   [arsmn/fiber-casbin](https://github.com/arsmn/fiber-casbin)
-   [arsmn/fiber-introspect](https://github.com/arsmn/fiber-introspect)
-   [shareed2k/fiber_tracing](https://github.com/shareed2k/fiber_tracing)
-   [shareed2k/fiber_limiter](https://github.com/shareed2k/fiber_limiter)
-   [thomasvvugt/fiber-boilerplate](https://github.com/thomasvvugt/fiber-boilerplate)
-   [arsmn/gqlgen](https://github.com/arsmn/gqlgen)
-   [kiyonlin/fiber_limiter](https://github.com/kiyonlin/fiber_limiter)
-   [juandiii/go-jwk-security](https://github.com/juandiii/go-jwk-security)
-   [sujit-baniya/fiber-boilerplate](https://github.com/sujit-baniya/fiber-boilerplate)
-   [ansrivas/fiberprometheus](https://github.com/ansrivas/fiberprometheus)
-   [LdDl/fiber-long-poll](https://github.com/LdDl/fiber-long-poll)
-   [K0enM/fiber_vhost](https://github.com/K0enM/fiber_vhost)
-   [sacsand/gofiber-firebaseauth](https://github.com/sacsand/gofiber-firebaseauth)
-   [theArtechnology/fiber-inertia](https://github.com/theArtechnology/fiber-inertia)
-   [aschenmaker/fiber-health-check](https://github.com/aschenmaker/fiber-health-check)
-   [elastic/apmfiber](https://github.com/elastic/apm-agent-go/tree/master/module/apmfiber)

## 👍 Contribute

If you want to say **thank you** and/or support the active development of `Fiber`:

1. Add a [GitHub Star](https://github.com/gofiber/fiber/stargazers) to the project.
2. Tweet about the project [on your Twitter](https://twitter.com/intent/tweet?text=Fiber%20is%20an%20Express%20inspired%20%23web%20%23framework%20built%20on%20top%20of%20Fasthttp%2C%20the%20fastest%20HTTP%20engine%20for%20%23Go.%20Designed%20to%20ease%20things%20up%20for%20%23fast%20development%20with%20zero%20memory%20allocation%20and%20%23performance%20in%20mind%20%F0%9F%9A%80%20https%3A%2F%2Fgithub.com%2Fgofiber%2Ffiber).
3. Write a review or tutorial on [Medium](https://medium.com/), [Dev.to](https://dev.to/) or personal blog.
4. Support the project by donating a [cup of coffee](https://buymeacoff.ee/fenny).

## ☕ Supporters

Fiber is an open source project that runs on donations to pay the bills e.g. our domain name, gitbook, netlify and serverless hosting. If you want to support Fiber, you can ☕ [**buy a coffee here**](https://buymeacoff.ee/fenny).

|                                                            | User                                             | Donation |
| :--------------------------------------------------------- | :----------------------------------------------- | :------- |
| ![](https://avatars.githubusercontent.com/u/204341?s=25)   | [@destari](https://github.com/destari)           | ☕ x 10  |
| ![](https://avatars.githubusercontent.com/u/63164982?s=25) | [@dembygenesis](https://github.com/dembygenesis) | ☕ x 5   |
| ![](https://avatars.githubusercontent.com/u/56607882?s=25) | [@thomasvvugt](https://github.com/thomasvvugt)   | ☕ x 5   |
| ![](https://avatars.githubusercontent.com/u/27820675?s=25) | [@hendratommy](https://github.com/hendratommy)   | ☕ x 5   |
| ![](https://avatars.githubusercontent.com/u/1094221?s=25)  | [@ekaputra07](https://github.com/ekaputra07)     | ☕ x 5   |
| ![](https://avatars.githubusercontent.com/u/194590?s=25)   | [@jorgefuertes](https://github.com/jorgefuertes) | ☕ x 5   |
| ![](https://avatars.githubusercontent.com/u/186637?s=25)   | [@candidosales](https://github.com/candidosales) | ☕ x 5   |
| ![](https://avatars.githubusercontent.com/u/29659953?s=25) | [@l0nax](https://github.com/l0nax)               | ☕ x 3   |
| ![](https://avatars.githubusercontent.com/u/59947262?s=25) | [@ankush](https://github.com/ankush)             | ☕ x 3   |
| ![](https://avatars.githubusercontent.com/u/635852?s=25)   | [@bihe](https://github.com/bihe)                 | ☕ x 3   |
| ![](https://avatars.githubusercontent.com/u/307334?s=25)   | [@justdave](https://github.com/justdave)         | ☕ x 3   |
| ![](https://avatars.githubusercontent.com/u/11155743?s=25) | [@koddr](https://github.com/koddr)               | ☕ x 1   |
| ![](https://avatars.githubusercontent.com/u/29042462?s=25) | [@lapolinar](https://github.com/lapolinar)       | ☕ x 1   |
| ![](https://avatars.githubusercontent.com/u/2978730?s=25)  | [@diegowifi](https://github.com/diegowifi)       | ☕ x 1   |
| ![](https://avatars.githubusercontent.com/u/44171355?s=25) | [@ssimk0](https://github.com/ssimk0)             | ☕ x 1   |
| ![](https://avatars.githubusercontent.com/u/5638101?s=25)  | [@raymayemir](https://github.com/raymayemir)     | ☕ x 1   |
| ![](https://avatars.githubusercontent.com/u/619996?s=25)   | [@melkorm](https://github.com/melkorm)           | ☕ x 1   |
| ![](https://avatars.githubusercontent.com/u/31022056?s=25) | [@marvinjwendt](https://github.com/marvinjwendt) | ☕ x 1   |
| ![](https://avatars.githubusercontent.com/u/31921460?s=25) | [@toishy](https://github.com/toishy)             | ☕ x 1   |

## ‎‍💻 Code Contributors

```
Placeholder
```
## ⭐️ Stargazers

```
Placeholder
```
## ⚠️ License

Copyright (c) 2019-present [Fenny](https://github.com/fenny) and [Contributors](https://github.com/gofiber/fiber/graphs/contributors). `Fiber` is free and open-source software licensed under the [MIT License](https://github.com/gofiber/fiber/blob/master/LICENSE). Official logo was created by [Vic Shóstak](https://github.com/koddr) and distributed under [Creative Commons](https://creativecommons.org/licenses/by-sa/4.0/) license (CC BY-SA 4.0 International).

**Third-party library licenses**

```
Placeholder
```
