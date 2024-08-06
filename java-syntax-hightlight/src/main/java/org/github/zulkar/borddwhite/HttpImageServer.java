package org.github.zulkar.borddwhite;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

public class HttpImageServer {
    private final HttpServer httpServer;
    private final ImageRenderer renderer;

    public HttpImageServer(int port, ImageRenderer renderer) throws IOException {
        this.renderer = renderer;
        httpServer = HttpServer.create();
        httpServer.bind(new InetSocketAddress(port), 0);
        httpServer.createContext("/", exchange -> {
            try {
                var params = queryMap(exchange);
                var lang = params.get("l");
                var code64 = params.get("c");
                var theme = params.get("t");
                if (theme == null || theme.isEmpty()) theme = "dark";
                var paddings = params.get("p");
                if (paddings == null || paddings.isEmpty()) paddings = "10";
                var code = new String(Base64.getDecoder().decode(code64), StandardCharsets.UTF_8);
                byte[] imageData = renderer.renderToPng(code, lang, theme, Integer.parseInt(paddings));
                exchange.getResponseHeaders().add("Content-Type", "image/png");
                exchange.sendResponseHeaders(200, imageData.length);
                exchange.getResponseBody().write(imageData);
                exchange.close();
            } catch (IllegalArgumentException e) {
                e.printStackTrace();
                exchange.sendResponseHeaders(400, 0);
                exchange.getResponseBody().write(e.getMessage().getBytes(StandardCharsets.UTF_8));
                exchange.close();
            } catch (Exception e) {
                e.printStackTrace();
                exchange.sendResponseHeaders(500, 0);
                exchange.getResponseBody().write(e.getMessage().getBytes(StandardCharsets.UTF_8));
                exchange.close();
            }
        });
    }

    public void start() {
        httpServer.start();
    }

    Map<String, String> queryMap(HttpExchange exchange) {
        var query = exchange.getRequestURI().getRawQuery();
        if (query == null || query.isEmpty()) {
            return Map.of();
        }
        Map<String, String> result = new HashMap<>();
        for (String param : query.split("&")) {
            String[] entry = param.split("=", 2);
            var name = entry[0];
            var value = entry.length > 1 ? URLDecoder.decode(entry[1], StandardCharsets.UTF_8) : "";
            result.put(name, value);
        }
        return result;
    }
}
