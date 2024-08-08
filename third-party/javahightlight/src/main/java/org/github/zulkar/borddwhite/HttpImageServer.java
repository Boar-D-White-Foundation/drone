package org.github.zulkar.borddwhite;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
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
    private final int port;

    public HttpImageServer(int port, ImageRenderer renderer) throws IOException {
        this.port = port;
        httpServer = HttpServer.create();
        httpServer.bind(new InetSocketAddress(port), 0);
        httpServer.createContext("/", new MyRenderHandler(renderer));
    }

    public void start() {
        httpServer.start();
        System.out.println("Render server started on port " + port);
    }


    private static class MyRenderHandler implements HttpHandler {

        private final ImageRenderer renderer;

        public MyRenderHandler(ImageRenderer renderer) {
            this.renderer = renderer;
        }

        Map<String, String> urlParamsMap(HttpExchange exchange) {
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

        @Override
        public void handle(HttpExchange exchange) throws IOException {

            switch (exchange.getRequestMethod()) {
                case "GET" -> withExceptionProcessing(exchange, this::processGet);
                case "POST" -> withExceptionProcessing(exchange, this::processPost);
            }

        }

        private void processPost(HttpExchange e) throws Exception {
            var params = urlParamsMap(e);
            var lang = params.get("l");
            var theme = params.get("t");
            if (theme == null || theme.isEmpty()) theme = "dark";
            var paddings = params.get("p");
            if (paddings == null || paddings.isEmpty()) paddings = "10";
            var code = new String(e.getRequestBody().readAllBytes(), StandardCharsets.UTF_8);
            byte[] imageData = renderer.renderToPng(code, lang, theme, Integer.parseInt(paddings));
            e.getResponseHeaders().add("Content-Type", "image/png");
            e.sendResponseHeaders(200, imageData.length);
            e.getResponseBody().write(imageData);
            e.close();
        }

        private void processGet(HttpExchange e) throws Exception {
            var params = urlParamsMap(e);
            var lang = params.get("l");
            var code64 = params.get("c64");
            var theme = params.get("t");
            if (theme == null || theme.isEmpty()) theme = "dark";
            var paddings = params.get("p");
            if (paddings == null || paddings.isEmpty()) paddings = "10";
            var code = new String(Base64.getDecoder().decode(code64), StandardCharsets.UTF_8);
            byte[] imageData = renderer.renderToPng(code, lang, theme, Integer.parseInt(paddings));
            e.getResponseHeaders().add("Content-Type", "image/png");
            e.sendResponseHeaders(200, imageData.length);
            e.getResponseBody().write(imageData);
            e.close();
        }

        private void withExceptionProcessing(HttpExchange exchange, RunThrowable run) throws IOException {
            try {
                run.run(exchange);
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
        }


    }

    private interface RunThrowable {
        void run(HttpExchange e) throws Exception;
    }
}

