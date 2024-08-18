package org.github.zulkar.borddwhite

import com.sun.net.httpserver.HttpExchange
import com.sun.net.httpserver.HttpHandler
import com.sun.net.httpserver.HttpServer
import java.net.InetSocketAddress
import java.net.URLDecoder
import java.nio.charset.StandardCharsets
import java.util.*
import java.util.Map

class HttpImageServer(private val port: Int, renderer: ImageRenderer) {
    private val httpServer: HttpServer = HttpServer.create()

    init {
        httpServer.bind(InetSocketAddress(port), 0)
        httpServer.createContext("/", MyRenderHandler(renderer))
    }

    fun start() {
        httpServer.start()
        println("Render server started on port $port")
    }

    private class MyRenderHandler(private val renderer: ImageRenderer) : HttpHandler {
        override fun handle(exchange: HttpExchange) {
            when (exchange.requestMethod) {
                "GET" -> withExceptionProcessing(exchange, this::processGet)
                "POST" -> withExceptionProcessing(exchange, this::processPost)
                else -> {
                    exchange.sendResponseHeaders(405, 0)
                    exchange.responseBody.write(
                        "${exchange.requestMethod} is not supported".toByteArray(
                            StandardCharsets.UTF_8
                        )
                    )
                    exchange.close()
                }
            }
        }

        fun processPost(e: HttpExchange) {
            val params = urlParamsMap(e)
            val lang = params["l"]
            var theme = params["t"]
            if (theme == null || theme.isEmpty()) theme = "dark"
            var paddings = params["p"]
            if (paddings == null || paddings.isEmpty()) paddings = "10"
            val code = String(e.requestBody.readAllBytes(), StandardCharsets.UTF_8)
            val useLigatures = params["ligatures"].toBoolean()
            val imageData = renderer.renderToPng(code, lang, theme, paddings.toInt(), useLigatures)
            e.responseHeaders.add("Content-Type", "image/png")
            e.sendResponseHeaders(200, imageData.size.toLong())
            e.responseBody.write(imageData)
            e.close()
        }

        fun processGet(e: HttpExchange) {
            val params = urlParamsMap(e)
            val lang = params["l"]
            val code64 = params["c64"]
            val codeP = params["c"]
            var theme = params["t"]
            if (theme == null || theme.isEmpty()) theme = "dark"
            var paddings = params["p"]
            if (paddings == null || paddings.isEmpty()) paddings = "10"
            val useLigatures = params["ligatures"].toBoolean()
            val code = codeP ?: String(Base64.getDecoder().decode(code64), StandardCharsets.UTF_8)
            val imageData = renderer.renderToPng(code, lang, theme, paddings.toInt(), useLigatures)
            e.responseHeaders.add("Content-Type", "image/png")
            e.sendResponseHeaders(200, imageData.size.toLong())
            e.responseBody.write(imageData)
            e.close()
        }

        fun withExceptionProcessing(exchange: HttpExchange, process: (HttpExchange) -> Unit) {
            try {
                process(exchange)
            } catch (e: IllegalArgumentException) {
                e.printStackTrace()
                exchange.sendResponseHeaders(400, 0)
                e.message?.let { exchange.responseBody.write(it.toByteArray(StandardCharsets.UTF_8)) }
                exchange.close()
            } catch (e: Exception) {
                e.printStackTrace()
                exchange.sendResponseHeaders(500, 0)
                e.message?.let { exchange.responseBody.write(it.toByteArray(StandardCharsets.UTF_8)) }
                exchange.close()
            }
        }
    }
}

private fun urlParamsMap(exchange: HttpExchange): MutableMap<String?, String?> {
    val query = exchange.requestURI.rawQuery
    if (query == null || query.isEmpty()) {
        return Map.of<String?, String?>()
    }
    val result: MutableMap<String?, String?> = HashMap<String?, String?>()
    for (param in query.split("&".toRegex()).dropLastWhile { it.isEmpty() }.toTypedArray()) {
        val entry: Array<String?> = param.split("=".toRegex(), limit = 2).toTypedArray()
        val name = entry[0]
        val value = if (entry.size > 1) URLDecoder.decode(entry[1], StandardCharsets.UTF_8) else ""
        result.put(name, value)
    }
    return result
}

