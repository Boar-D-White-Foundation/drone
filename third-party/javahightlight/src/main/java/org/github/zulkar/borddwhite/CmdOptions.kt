package org.github.zulkar.borddwhite

import picocli.CommandLine
import java.io.File

class CmdOptions {
    @CommandLine.Option(
        names = ["-t", "--theme"],
        paramLabel = "THEME",
        defaultValue = "dark",
        description = ["theme to use"]
    )
    var theme: String? = null

    @CommandLine.Option(names = ["-l", "--lang"], paramLabel = "LANG", description = ["Language"])
    var lang: String? = null

    @CommandLine.Option(
        names = ["-i", "--input"],
        paramLabel = "INPUT",
        description = ["input text file"],
        required = true
    )
    lateinit var input: File

    @CommandLine.Option(
        names = ["-o", "--output"],
        paramLabel = "OUTPUT",
        description = ["output image file"],
        required = true
    )
    lateinit var output: File

    @CommandLine.Option(names = ["-p", "--paddings"], paramLabel = "PADDINGS", description = ["padding around image"])
    var paddings: Int = 0
}
