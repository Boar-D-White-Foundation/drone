package org.github.zulkar.borddwhite;

import org.fife.ui.rsyntaxtextarea.RSyntaxTextArea;
import org.fife.ui.rsyntaxtextarea.SyntaxConstants;
import org.fife.ui.rsyntaxtextarea.Theme;

import javax.imageio.ImageIO;
import javax.swing.*;
import java.awt.*;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.util.Collections;
import java.util.Comparator;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.atomic.AtomicReference;

public class ImageRenderer {
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


    private final ConcurrentMap<String, Theme> themes;

    public ImageRenderer() {
        themes = new ConcurrentHashMap<>();
    }

    public void initialize() throws Exception {
        loadFonts();
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

    public  byte[] renderToPng(String code, String lang, String themeName, int paddings) throws Exception {
        var font = new Font("JetBrains Mono", Font.PLAIN, 30);
        var theme = getOrLoadTheme(themeName);
        var textArea = prepareRSyntax(code, lang, theme, font);
        return render(textArea, paddings);
    }

    private Theme getOrLoadTheme(String themeName) {
        var theme = themes.get(themeName); //no need computeIfAbsent,in case of race condition just load twice. Better stay robust if theme does not exist

        if (theme == null) {
            try {
                theme = loadTheme(themeName);
                themes.put(themeName, theme);
            } catch (IOException ignore) {

            }
        }
        return theme;
    }

    private static RSyntaxTextArea prepareRSyntax(String code, String lang, Theme theme, Font font) {
        var lines = code.lines().toList();
        RSyntaxTextArea textArea = new RSyntaxTextArea(lines.size(), Collections.max(lines, Comparator.comparing(String::length)).length());
        if (theme != null) {
            theme.apply(textArea);
        }
        textArea.setSyntaxEditingStyle(getLanguage(lang));
        textArea.setCodeFoldingEnabled(true);
        textArea.setHighlightCurrentLine(false);
        textArea.setAntiAliasingEnabled(true);
        textArea.setFont(font);
        textArea.setAntiAliasingEnabled(true);
        textArea.setBracketMatchingEnabled(true);
        textArea.setText(code);
        textArea.getCaret().setSelectionVisible(false);
        textArea.getCaret().setVisible(false);
        textArea.setEditable(false);
        return textArea;
    }

    private byte[] render(RSyntaxTextArea textArea, int paddings) throws Exception {
        AtomicReference<byte[]> ref = new AtomicReference<>(new byte[0]);
        SwingUtilities.invokeAndWait(() -> {
            var frame = new JPanel();
            frame.setVisible(true);
            frame.add(textArea);
            textArea.addNotify();
            frame.setSize(frame.getPreferredSize());
            textArea.setBounds(0, 0, frame.getWidth(), frame.getHeight());
            frame.getGraphics();

            BufferedImage im = new BufferedImage(textArea.getWidth(), textArea.getHeight(), BufferedImage.TYPE_INT_ARGB);
            Graphics2D renderGraphics = im.createGraphics();
            refreshMetrics(textArea, renderGraphics);
            frame.paint(renderGraphics);
            BufferedImage imageWithPaddings = new BufferedImage(textArea.getWidth() + paddings * 2, textArea.getHeight() + paddings * 2, BufferedImage.TYPE_INT_ARGB);

            Graphics2D graphics = imageWithPaddings.createGraphics();
            graphics.setColor(textArea.getBackground());
            graphics.fillRect(0, 0, imageWithPaddings.getWidth(), imageWithPaddings.getHeight());
            graphics.drawImage(im, paddings, paddings, null);
            ByteArrayOutputStream output = new ByteArrayOutputStream();
            try {
                ImageIO.write(imageWithPaddings, "PNG", output);
                ref.set(output.toByteArray());
            } catch (IOException e) {
                throw new RuntimeException(e);
            }

        });
        return ref.get();
    }

    private void refreshMetrics(RSyntaxTextArea textArea, Graphics2D g) {
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

    public static String getLanguage(String lang) {
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
}
