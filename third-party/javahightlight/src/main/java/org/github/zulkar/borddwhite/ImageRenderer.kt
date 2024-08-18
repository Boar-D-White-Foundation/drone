package org.github.zulkar.borddwhite

import org.fife.ui.rsyntaxtextarea.RSyntaxTextArea
import org.fife.ui.rsyntaxtextarea.SyntaxConstants
import org.fife.ui.rsyntaxtextarea.Theme
import java.awt.Font
import java.awt.Graphics2D
import java.awt.GraphicsEnvironment
import java.awt.font.TextAttribute
import java.awt.font.TextAttribute.LIGATURES
import java.awt.image.BufferedImage
import java.io.ByteArrayOutputStream
import java.io.IOException
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.ConcurrentMap
import java.util.concurrent.atomic.AtomicReference
import javax.imageio.ImageIO
import javax.swing.JPanel
import javax.swing.SwingUtilities
import kotlin.streams.asStream

class ImageRenderer {
    private var initialized = false
    private lateinit var withLigatures: Font
    private lateinit var withoutLigatures: Font


    private val themes: ConcurrentMap<String?, Theme?> = ConcurrentHashMap<String?, Theme?>()

    fun initialize() {
        loadFonts()
        withoutLigatures = Font("JetBrains Mono", Font.PLAIN, 30)

        @Suppress("UNCHECKED_CAST")
        val attributes = withoutLigatures.attributes as MutableMap<TextAttribute, Any>
        attributes[LIGATURES] = TextAttribute.LIGATURES_ON
        withLigatures = withoutLigatures.deriveFont(attributes)
        initialized = true
    }

    private fun loadTheme(theme: String?): Theme? {
        return Theme.load(
            javaClass.getResourceAsStream("/themes/rsyntax/$theme.xml")
        )
    }

    private fun loadFonts() {
        val graphicsEnvironment = GraphicsEnvironment.getLocalGraphicsEnvironment()
        val availableFontFamilyNames = graphicsEnvironment.availableFontFamilyNames
        for (name in availableFontFamilyNames) {
            if ("JetBrains Mono" == name) return
        }

        for (fontName in FONT_NAMES) {
            this.javaClass.getResourceAsStream("/fonts/JetBrainsMono/$fontName").use { stream ->
                val font = Font.createFont(Font.TRUETYPE_FONT, stream)
                graphicsEnvironment.registerFont(font)
            }
        }
    }

    fun renderToPng(code: String, lang: String?, themeName: String, paddings: Int, useLigatures: Boolean): ByteArray {
        var code = code
        check(initialized) { "call initialize first" }
        code = removeFuckingTabs(code)

        val theme = getOrLoadTheme(themeName)
        val textArea: RSyntaxTextArea =
            prepareRSyntax(code, lang, theme, if (useLigatures) withLigatures else withoutLigatures)
        return render(textArea, paddings)
    }

    private fun removeFuckingTabs(code: String): String {
        return code.replace("\t", "    ")
    }

    private fun getOrLoadTheme(themeName: String): Theme? {
        var theme =
            themes[themeName] //no need computeIfAbsent, in case of race condition just load twice. Better stay robust if the theme does not exist

        if (theme == null) {
            try {
                theme = loadTheme(themeName)
                themes.put(themeName, theme)
            } catch (_: IOException) {
            }
        }
        return theme
    }

    private fun render(textArea: RSyntaxTextArea, paddings: Int): ByteArray {
        val ref = AtomicReference<ByteArray>(ByteArray(0))
        SwingUtilities.invokeAndWait(Runnable {
            val frame = JPanel()
            frame.isVisible = true
            frame.add(textArea)
            textArea.addNotify()
            frame.size = frame.getPreferredSize()
            textArea.setBounds(0, 0, frame.getWidth(), frame.getHeight())
            frame.getGraphics()

            val im = BufferedImage(textArea.getWidth(), textArea.getHeight(), BufferedImage.TYPE_INT_ARGB)
            val renderGraphics = im.createGraphics()
            refreshMetrics(textArea, renderGraphics)
            frame.paint(renderGraphics)
            val imageWithPaddings = BufferedImage(
                textArea.getWidth() + paddings * 2,
                textArea.getHeight() + paddings * 2,
                BufferedImage.TYPE_INT_ARGB
            )

            val graphics = imageWithPaddings.createGraphics()
            graphics.color = textArea.getBackground()
            graphics.fillRect(0, 0, imageWithPaddings.width, imageWithPaddings.height)
            graphics.drawImage(im, paddings, paddings, null)
            val output = ByteArrayOutputStream()
            try {
                ImageIO.write(imageWithPaddings, "PNG", output)
                ref.set(output.toByteArray())
            } catch (e: IOException) {
                throw RuntimeException(e)
            }
        })
        return ref.get()
    }

    private fun refreshMetrics(textArea: RSyntaxTextArea?, g: Graphics2D?) {
        try {
            val clazz: Class<*> = RSyntaxTextArea::class.java
            val f = clazz.getDeclaredField("metricsNeverRefreshed")
            f.setAccessible(true)
            f.setBoolean(textArea, false)
            val m = clazz.getDeclaredMethod("refreshFontMetrics", Graphics2D::class.java)
            m.setAccessible(true)
            m.invoke(textArea, g)
        } catch (e: ReflectiveOperationException) {
            throw RuntimeException(e)
        }
    }

    companion object {
        val FONT_NAMES: Array<String> = arrayOf(
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
            "JetBrainsMono-ThinItalic.ttf"
        )

        private fun prepareRSyntax(code: String, lang: String?, theme: Theme?, font: Font?): RSyntaxTextArea {
            val lines = code.lineSequence().asStream().toList()
            val textArea = RSyntaxTextArea(
                lines.size, lines.maxOf { it.length }
            )
            theme?.apply(textArea)
            textArea.setSyntaxEditingStyle(getLanguage(lang))
            textArea.isCodeFoldingEnabled = true
            textArea.highlightCurrentLine = false
            textArea.setAntiAliasingEnabled(true)
            textArea.setFont(font)
            textArea.setAntiAliasingEnabled(true)
            textArea.isBracketMatchingEnabled = true
            textArea.text = code
            textArea.caret.isSelectionVisible = false
            textArea.caret.isVisible = false
            textArea.isEditable = false
            return textArea
        }

        fun getLanguage(lang: String?): String? {
            if (lang == null) return null
            when (lang) {
                "cpp" -> return SyntaxConstants.SYNTAX_STYLE_CPLUSPLUS
                "java" -> return SyntaxConstants.SYNTAX_STYLE_JAVA
                "python", "python3" -> return SyntaxConstants.SYNTAX_STYLE_PYTHON
                "c" -> return SyntaxConstants.SYNTAX_STYLE_C
                "csharp" -> return SyntaxConstants.SYNTAX_STYLE_CSHARP
                "javascript" -> return SyntaxConstants.SYNTAX_STYLE_JAVASCRIPT
                "typescript" -> return SyntaxConstants.SYNTAX_STYLE_TYPESCRIPT
                "php" -> return SyntaxConstants.SYNTAX_STYLE_PHP
                "swift" -> return null
                "kotlin" -> return SyntaxConstants.SYNTAX_STYLE_KOTLIN
                "golang" -> return SyntaxConstants.SYNTAX_STYLE_GO
                "ruby" -> return SyntaxConstants.SYNTAX_STYLE_RUBY
                "scala" -> return SyntaxConstants.SYNTAX_STYLE_SCALA
                "rust" -> return SyntaxConstants.SYNTAX_STYLE_RUST
                "racket" -> return null
                else -> return null
            }
        }
    }
}
