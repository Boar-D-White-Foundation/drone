package org.github.zulkar.borddwhite;

import org.junit.jupiter.api.*;

import java.nio.charset.StandardCharsets;

class ImageRendererTest {

    @BeforeEach
    void setUp() {
        System.setProperty("java.awt.headless", "true");

    }

    @AfterEach
    void tearDown() {
    }

    @Test
    void renderToPng() throws Exception {
        ImageRenderer imageRenderer = new ImageRenderer();
        imageRenderer.initialize();
        Assumptions.assumeTrue(java.awt.GraphicsEnvironment.isHeadless());
        try (var inp = getClass().getResourceAsStream("/input/input.java");
             var exp = getClass().getResourceAsStream("/output/java.png")) {
            String text = new String(inp.readAllBytes(), StandardCharsets.UTF_8);
            byte[] data = imageRenderer.renderToPng(text, "java", "dark", 20);
            Assertions.assertArrayEquals(exp.readAllBytes(), data);
        }
    }
}