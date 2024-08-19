package org.github.zulkar.borddwhite

import kotlin.system.exitProcess

object HttpServerMain {
    @JvmStatic
    fun main(args: Array<String>) {
        if (args.isEmpty()) {
            System.err.println("First parameter should be port number")
            exitProcess(1)
        }
        val port = args[0].toInt()
        val renderer = ImageRenderer()
        renderer.initialize()
        val server = HttpImageServer(port, renderer)
        server.start()
    }
}
