'use strict'

const path = require('path')
const AutoLoad = require('@fastify/autoload')
const promClient = require('prom-client')

module.exports = async function (fastify, opts) {
// Place here your custom code!

  // Create a Registry which registers the metrics
  const register = new promClient.Registry()
  
  // Add a default label which is added to all metrics
  register.setDefaultLabels({
    app: 'order-service'
  })
  
  // Enable the collection of default metrics
  promClient.collectDefaultMetrics({
    register
  })
  
  // Add metrics endpoint
  fastify.get('/metrics', async (request, reply) => {
    reply.type('text/plain')
    return register.metrics()
  })

  fastify.register(require('@fastify/cors'), {
    origin: '*'
  })

  // Do not touch the following lines

  // This loads all plugins defined in plugins
  // those should be support plugins that are reused
  // through your application
  fastify.register(AutoLoad, {
    dir: path.join(__dirname, 'plugins'),
    options: Object.assign({}, opts)
  })

  // This loads all plugins defined in routes
  // define your routes in one of these
  fastify.register(AutoLoad, {
    dir: path.join(__dirname, 'routes'),
    options: Object.assign({}, opts)
  })
}
