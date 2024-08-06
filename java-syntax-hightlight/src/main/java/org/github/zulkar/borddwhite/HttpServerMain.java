package org.github.zulkar.borddwhite;

public class HttpServerMain {
    public static void main(String[] args) throws Exception {
        if (args.length == 0) {
            System.err.println("First parameter should be port number");
            System.exit(1);
        }
        var port = Integer.parseInt(args[0]);
        var renderer = new ImageRenderer();
        renderer.initialize();
        HttpImageServer server = new HttpImageServer(port, renderer);
        server.start();
    }
}
