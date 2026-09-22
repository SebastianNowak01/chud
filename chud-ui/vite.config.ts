import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import { defineConfig } from 'vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import { fileURLToPath, URL } from 'node:url'
import { resolve } from 'node:path'
import { cp, mkdir, rm } from 'node:fs/promises'

const uiDistDir = 'dist'
const serverStaticDir = fileURLToPath(
  new URL('../chud-server/src/static', import.meta.url),
)

const copyServerStatic = () => ({
  name: 'copy-server-static',
  async closeBundle() {
    const sourceDir = resolve(process.cwd(), uiDistDir)

    await rm(serverStaticDir, { recursive: true, force: true })
    await mkdir(serverStaticDir, { recursive: true })
    await cp(sourceDir, serverStaticDir, { recursive: true })
  },
})

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
    }),
    react(),
    babel({ presets: [reactCompilerPreset()] }),
    copyServerStatic(),
  ],
  build: {
    outDir: uiDistDir,
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:2137',
    },
  },
})
