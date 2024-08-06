package org.github.zulkar.borddwhite;

import picocli.CommandLine;

import java.io.File;

public class CmdOptions {
    @CommandLine.Option(names = {"-t", "--theme"}, paramLabel = "THEME", defaultValue = "dark", description = "theme to use")
    String theme;

    @CommandLine.Option(names = {"-l", "--lang"}, paramLabel = "LANG", description = "Language")
    String lang;

    @CommandLine.Option(names = {"-i", "--input"}, paramLabel = "INPUT", description = "input text file", required = true)
    File input;

    @CommandLine.Option(names = {"-o", "--output"}, paramLabel = "OUTPUT", description = "output image file", required = true)
    File output;

    @CommandLine.Option(names = {"-p", "--paddings"}, paramLabel = "PADDINGS", description = "padding around image")
    int paddings;
}
