package org.github.zulkar.borddwhite

import picocli.CommandLine
import picocli.CommandLine.ParameterException
import java.io.File
import java.nio.file.Files
import kotlin.system.exitProcess

class ImagePreviewGenerator {
    private val renderer: ImageRenderer = ImageRenderer()

    init {
        renderer.initialize()
    }

    fun generateImage(options: CmdOptions) {
        val code = loadFile(options.input)
        val data = renderer.renderToPng(code, options.lang, options.theme, options.paddings, false)
        Files.write(options.output.toPath(), data)
    }

    private fun loadFile(input: File): String? {
        return Files.readString(input.toPath())
    }
}

fun main(args: Array<String>) {
    val options = CmdOptions()
    val cmd = CommandLine(options)
    try {
        val parseResult = cmd.parseArgs(*args)
        if (parseResult.isUsageHelpRequested) {
            cmd.usage(cmd.out)
            return
        }
        ImagePreviewGenerator().generateImage(options)
    } catch (ex: ParameterException) {
        cmd.err.println(ex.message)
        if (!CommandLine.UnmatchedArgumentException.printSuggestions(ex, cmd.getErr())) {
            ex.getCommandLine().usage(cmd.err)
        }
        exitProcess(1)
    } catch (e: Exception) {
        e.printStackTrace()
        exitProcess(2)
    }
}
