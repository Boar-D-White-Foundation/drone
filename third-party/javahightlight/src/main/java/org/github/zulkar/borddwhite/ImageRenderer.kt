package org.github.zulkar.borddwhite

import org.fife.ui.rsyntaxtextarea.RSyntaxTextArea
import org.fife.ui.rsyntaxtextarea.SyntaxConstants
import org.fife.ui.rsyntaxtextarea.Theme
import java.awt.Font
import java.awt.Graphics2D
import java.awt.GraphicsEnvironment
import java.awt.Image.SCALE_SMOOTH
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


    private val themes: ConcurrentMap<String?, Theme?> = ConcurrentHashMap<String?, Theme?>()

    fun initialize() {
        loadFonts()

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
        var min = 18
        var max = 30

        val start = System.currentTimeMillis()
        val optimistic = getFont(max, useLigatures)
        val firstResult = doRenderToPng(code, lang, themeName, optimistic, paddings)
        if (firstResult.w <= MAX && firstResult.h <= MAX) {
            val image = firstResult.image.toPng()
            println("generated image with first try. Size=${image.size} bytes ${firstResult.w}x${firstResult.h} with fontSize = $max and in ${System.currentTimeMillis() - start}ms")
            return image
        }

        if ((firstResult.w >= MAX * 1.5 && firstResult.w <= MAX * 2.0) || (firstResult.h >= MAX * 1.5 && firstResult.h <= MAX * 2.0)) {
            val image = firstResult.image.scale().toPng()
            println("generated image and scale to 0.5. Size=${image.size} bytes ${firstResult.w / 2}x${firstResult.h / 2} with fontSize = $max and in ${System.currentTimeMillis() - start}ms")
            return image
        }

        var ans = firstResult
        var counter = 1
        while (min < max) {
            val mid = (max + min) / 2
            counter++
            val result = doRenderToPng(code, lang, themeName, getFont(mid, useLigatures), paddings)
            if (result.w <= MAX && result.h <= MAX) {
                ans = result
                min = mid + 1
            } else {
                max = mid - 1
            }
        }
        val image = ans.image.toPng()
        println("generated image in $counter tries with size=${image.size} bytes ${ans.w}x${ans.h} with fontSize = ${ans.font} and in ${System.currentTimeMillis() - start}ms")
        return image
    }

    private fun getFont(size: Int, useLigatures: Boolean): Font {
        val font = Font("JetBrains Mono", Font.PLAIN, size)
        if (!useLigatures) return font
        @Suppress("UNCHECKED_CAST")
        val attributes = font.attributes as MutableMap<TextAttribute, Any>
        attributes[LIGATURES] = TextAttribute.LIGATURES_ON
        return font.deriveFont(attributes)
    }

    private fun doRenderToPng(
        code: String,
        lang: String?,
        themeName: String,
        font: Font,
        paddings: Int
    ): RenderResult {
        var code = code
        check(initialized) { "call initialize first" }
        code = removeFuckingTabs(code)

        val theme = getOrLoadTheme(themeName, font)
        val textArea: RSyntaxTextArea =
            prepareRSyntax(code, lang, theme, font)
        return render(textArea, paddings)
    }

    private fun removeFuckingTabs(code: String): String {
        return code.replace("\t", "    ")
    }

    private fun getOrLoadTheme(themeName: String, font: Font): Theme? {
        val key = "$themeName|${font.name}|${font.size}"
        var theme: Theme? =
            themes[key] //no need computeIfAbsent, in case of race condition just load twice. Better stay robust if the theme does not exist

        if (theme == null) {
            try {
                theme = loadTheme(themeName)
                themes.put(key, theme)
            } catch (_: IOException) {
            }
        }
        return theme
    }

    private fun render(textArea: RSyntaxTextArea, paddings: Int): RenderResult {
        val ref = AtomicReference<RenderResult>()
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
            ref.set(
                RenderResult(
                    imageWithPaddings.width,
                    imageWithPaddings.height,
                    textArea.font.size,
                    imageWithPaddings
                )
            )
            renderGraphics.dispose()
            graphics.dispose()
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
        const val MAX = 2560
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


        private fun prepareRSyntax(code: String, lang: String?, theme: Theme?, font: Font): RSyntaxTextArea {
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

private fun BufferedImage.scale(): BufferedImage {
    val img = this.getScaledInstance(width / 2, height / 2, SCALE_SMOOTH)
    if (img is BufferedImage) return img
    val res = BufferedImage(img.getWidth(null), img.getHeight(null), BufferedImage.TYPE_INT_ARGB)
    val g2 = res.createGraphics();
    g2.drawImage(img, 0, 0, null);
    g2.dispose();
    return res
}

private fun BufferedImage.toPng(): ByteArray {
    val output = ByteArrayOutputStream()
    ImageIO.write(this, "PNG", output)
    return output.toByteArray()
}

private class RenderResult(val w: Int, val h: Int, val font: Int, val image: BufferedImage)
