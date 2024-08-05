package org.github.zulkar.borddwhite;

import org.fife.ui.rsyntaxtextarea.RSyntaxTextArea;
import org.fife.ui.rsyntaxtextarea.SyntaxConstants;
import org.fife.ui.rsyntaxtextarea.Theme;
import picocli.CommandLine;

import javax.imageio.ImageIO;
import javax.swing.*;
import java.awt.*;
import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.nio.file.Files;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;

public class ImagePreviewGenerator {

    public static final String[] FONT_NAMES = new String[]{
            "JetBrainsMono-Bold.ttf",
            "JetBrainsMono-BoldItalic.ttf",
            "JetBrainsMono-ExtraBold.ttf",
            "JetBrainsMono-ExtraBoldItalic.ttf",
            "JetBrainsMono-ExtraLight.ttf",
            "JetBrainsMono-ExtraLightItalic.ttf",
            "JetBrainsMono-Italic.ttf",
            "JetBrainsMono-Light.ttf",
            "JetBrainsMono-LightItalic.ttf",
            "JetBrainsMono-Medium.ttf",
            "JetBrainsMono-MediumItalic.ttf",
            "JetBrainsMono-Regular.ttf",
            "JetBrainsMono-SemiBold.ttf",
            "JetBrainsMono-SemiBoldItalic.ttf",
            "JetBrainsMono-Thin.ttf",
            "JetBrainsMono-ThinItalic.ttf"};

    public ImagePreviewGenerator() throws IOException, FontFormatException {
        loadFonts();
    }

    public void generateImage(CmdOptions options) throws IOException {
        var font = new Font("JetBrains Mono", Font.PLAIN, 30);
        var theme = loadTheme(options.theme);
        var lines = loadFile(options.input);
        RSyntaxTextArea textArea = new RSyntaxTextArea(lines.size(), Collections.max(lines, Comparator.comparing(String::length)).length());
        var text = new StringBuilder();
        for (String line : lines) {
            text.append(line).append(System.lineSeparator());
        }
        theme.apply(textArea);
        textArea.setSyntaxEditingStyle(getLanguage(options.lang));
        textArea.setCodeFoldingEnabled(true);
        textArea.setHighlightCurrentLine(false);
        textArea.setAntiAliasingEnabled(true);
        Font prev = textArea.getFont();
        textArea.setFont(font);
        textArea.setAntiAliasingEnabled(true);
        textArea.setBracketMatchingEnabled(true);
        textArea.setText(text.toString());
        textArea.getCaret().setSelectionVisible(false);
        textArea.getCaret().setVisible(false);
        textArea.setEditable(false);
        saveToImage(textArea, options.output, options.paddings);
    }

    private void saveToImage(RSyntaxTextArea textArea, File output, int padding) throws IOException {
        SwingUtilities.invokeLater(() -> {
            try {
                var frame = new JPanel();
                frame.setVisible(true);
                frame.add(textArea);
                textArea.addNotify();
                frame.setSize(frame.getPreferredSize());
                textArea.setBounds(0, 0, frame.getWidth(), frame.getHeight());
                frame.getGraphics();

                BufferedImage im = new BufferedImage(textArea.getWidth(), textArea.getHeight(), BufferedImage.TYPE_INT_ARGB);
                Graphics2D renderGraphics = im.createGraphics();
                setNoMetrics(textArea, renderGraphics);
                frame.paint(renderGraphics);
                BufferedImage imageWithPaddings = new BufferedImage(textArea.getWidth() + padding * 2, textArea.getHeight() + padding * 2, BufferedImage.TYPE_INT_ARGB);

                Graphics2D graphics = imageWithPaddings.createGraphics();
                graphics.setColor(textArea.getBackground());
                graphics.fillRect(0, 0, imageWithPaddings.getWidth(), imageWithPaddings.getHeight());
                graphics.drawImage(im, padding, padding, null);
                ImageIO.write(imageWithPaddings, "PNG", output);
                System.exit(0);
            } catch (IOException e) {
                e.printStackTrace();
                System.exit(1);
            }
        });


    }

    private void setNoMetrics(RSyntaxTextArea textArea, Graphics2D g) {
        try {
            Class<?> clazz = RSyntaxTextArea.class;
            Field f = clazz.getDeclaredField("metricsNeverRefreshed");
            f.setAccessible(true);
            f.setBoolean(textArea, false);
            Method m = clazz.getDeclaredMethod("refreshFontMetrics", Graphics2D.class);
            m.setAccessible(true);
            m.invoke(textArea, g);
        } catch (ReflectiveOperationException e) {
            throw new RuntimeException(e);
        }
    }

    private String getLanguage(String lang) {
        if (lang == null) return null;
        switch (lang) {
            case "cpp":
                return SyntaxConstants.SYNTAX_STYLE_CPLUSPLUS;
            case "java":
                return SyntaxConstants.SYNTAX_STYLE_JAVA;
            case "python":
            case "python3":
                return SyntaxConstants.SYNTAX_STYLE_PYTHON;
            case "c":
                return SyntaxConstants.SYNTAX_STYLE_C;
            case "csharp":
                return SyntaxConstants.SYNTAX_STYLE_CSHARP;
            case "javascript":
                return SyntaxConstants.SYNTAX_STYLE_JAVASCRIPT;
            case "typescript":
                return SyntaxConstants.SYNTAX_STYLE_TYPESCRIPT;
            case "php":
                return SyntaxConstants.SYNTAX_STYLE_PHP;
            case "swift":
                return null;
            case "kotlin":
                return SyntaxConstants.SYNTAX_STYLE_KOTLIN;
            case "golang":
                return SyntaxConstants.SYNTAX_STYLE_GO;
            case "ruby":
                return SyntaxConstants.SYNTAX_STYLE_RUBY;
            case "scala":
                return SyntaxConstants.SYNTAX_STYLE_SCALA;
            case "rust":
                return SyntaxConstants.SYNTAX_STYLE_RUST;
            case "racket":
                return null;
            default:
                return null;
        }
    }

    private List<String> loadFile(File input) throws IOException {
        return Files.readAllLines(input.toPath());
    }

    private Theme loadTheme(String theme) throws IOException {
        return Theme.load(getClass().getResourceAsStream(
                "/org/fife/ui/rsyntaxtextarea/themes/" + theme + ".xml"));
    }

    private void loadFonts() throws IOException, FontFormatException {
        final GraphicsEnvironment GE = GraphicsEnvironment.getLocalGraphicsEnvironment();
        String[] availableFontFamilyNames = GE.getAvailableFontFamilyNames();
        for (String name : availableFontFamilyNames) {
            if ("JetBrains Mono".equals(name)) return;
        }

        for (String fontName : FONT_NAMES) {
            try (InputStream stream = this.getClass().getResourceAsStream("/fonts/JetBrainsMono/" + fontName)) {
                Font font = Font.createFont(Font.TRUETYPE_FONT, stream);
                GE.registerFont(font);
            }
        }
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
