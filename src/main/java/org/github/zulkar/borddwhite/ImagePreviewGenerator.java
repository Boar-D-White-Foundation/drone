package org.github.zulkar.borddwhite;

import picocli.CommandLine;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;

public class ImagePreviewGenerator {
    private final ImageRenderer renderer;

    public ImagePreviewGenerator() throws Exception {
        renderer = new ImageRenderer();
        renderer.initialize();
    }

    public void generateImage(CmdOptions options) throws Exception {
        var code = loadFile(options.input);
        byte[] data = renderer.renderToPng(code, options.lang, options.theme, options.paddings);
        Files.write(options.output.toPath(), data);
    }

    private String loadFile(File input) throws IOException {
        return Files.readString(input.toPath());
    }


    public static void main(String[] args) throws IOException {
        CmdOptions options = new CmdOptions();
        CommandLine cmd = new CommandLine(options);
        try {
            CommandLine.ParseResult parseResult = cmd.parseArgs(args);
            if (parseResult.isUsageHelpRequested()) {
                cmd.usage(cmd.getOut());
                return;
            }
            new ImagePreviewGenerator().generateImage(options);
        } catch (CommandLine.ParameterException ex) {
            cmd.getErr().println(ex.getMessage());
            if (!CommandLine.UnmatchedArgumentException.printSuggestions(ex, cmd.getErr())) {
                ex.getCommandLine().usage(cmd.getErr());
            }
            System.exit(1);
        } catch (Exception e) {
            e.printStackTrace();
            System.exit(2);

        }
    }
}
