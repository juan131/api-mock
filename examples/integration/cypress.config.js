const { defineConfig } = require('cypress')

module.exports = defineConfig({
  e2e: {
    baseUrl: 'http://api-mock:8080',
    screenshotOnRunFailure: false,
    supportFile: false,
    video: false,
  }
})
