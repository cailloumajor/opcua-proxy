const subscriptionInterval = 500

describe("Integration tests page", () => {
  before(() => {
    cy.visit("index.html")
    cy.window().its("Centrifuge").should("be.a", "function")
    cy.request({ url: Cypress.env("HEALTH_URL"), retryOnStatusCodeFailure: true })
  })

  it("connects", () => {
    cy.get("#url-input").type(Cypress.env("CENTRIFUGO_URL"))
    cy.get("#connect-button").click()
    cy.get("#connect-status").should("have.text", "connected")
  })

  it("subscribes", () => {
    const ns = Cypress.env("CENTRIFUGO_NAMESPACE")
    const data = {
      namespaceURI: "urn:open62541.server.application",
      nodes: ["the.answer", 2345]
    }
    cy.get("#channel-input").type(`${ns}:integration@${subscriptionInterval}`)
    cy.get("#data-input")
      .clear()
      .type(JSON.stringify(data), { parseSpecialCharSequences: false })
    cy.get("#subscribe-button").click()
    cy.get("#subscription-status").should("have.text", "subscribed")
    let initial_date
    cy.get("#publication").children("li")
      .should(($li) => {
        expect($li).to.have.length(2)
        expect($li[0].innerText).to.equal("42")
        initial_date = new Date($li[1].innerText)
        expect(initial_date).to.be.above(new Date(0))
      })
      .wait(subscriptionInterval)
      .should(($li) => {
        expect($li).to.have.length(2)
        expect($li[0].innerText).to.equal("42")
        expect(new Date($li[1].innerText)).to.be.above(initial_date)
      })
  })

  it("unsubscribes", () => {
    cy.get("#unsubscribe-button").click()
    cy.get("#subscription-status").should("have.text", "unsubscribed")
    let current_time
    cy.get(("#publication")).children("li")
      .then(($li) => {
        current_time = $li[1].innerText
      })
      .wait(subscriptionInterval)
      .should(($li) => {
        expect($li).to.have.length(2)
        expect($li[0].innerText).to.equal("42")
        expect($li[1].innerText).to.equal(current_time)
      })
  })
})
