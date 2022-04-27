const subscriptionInterval = 500

before(() => {
  cy.request({
    url: `${Cypress.env("PROXY_URL")}/health`,
    retryOnStatusCodeFailure: true,
  })
})

describe("Integration tests page", () => {
  before(() => {
    cy.visit("index.html")
    cy.window().its("Centrifuge").should("be.a", "function")
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
      nodes: ["the.answer", 2345],
    }
    cy.get("#channel-input").type(`${ns}:integration@${subscriptionInterval}`)
    cy.get("#data-input")
      .clear()
      .type(JSON.stringify(data), { parseSpecialCharSequences: false })
    cy.get("#subscribe-button").click()
    cy.get("#subscription-status").should("have.text", "subscribed")
    let initialDate
    cy.get("#publication")
      .children("li")
      .as("publications")
      .should("have.length", 2)
      .its(0)
      .should("have.text", "42")
    cy.get("@publications").should(($li) => {
      initialDate = new Date($li[1].innerText)
      expect(initialDate).to.be.above(new Date(0))
    })
    cy.wait(subscriptionInterval)
    cy.get("@publications")
      .should("have.length", 2)
      .its(0)
      .should("have.text", "42")
    cy.get("@publications").should(($li) => {
      expect(new Date($li[1].innerText)).to.be.above(initialDate)
    })
  })

  it("unsubscribes", () => {
    cy.get("#unsubscribe-button").click()
    cy.get("#subscription-status").should("have.text", "unsubscribed")
    let currentTime
    cy.get("#publication")
      .children("li")
      .as("publications")
      .then(($li) => {
        currentTime = $li[1].innerText
      })
    cy.wait(subscriptionInterval)
    cy.get("@publications")
      .should("have.length", 2)
      .its(0)
      .should("have.text", "42")
    cy.get("@publications")
      .its(1)
      .should(($li) => {
        expect($li).to.have.text(currentTime)
      })
  })
})

describe("Nodes values endpoint", () => {
  it("returns valid data", () => {
    cy.request(`${Cypress.env("PROXY_URL")}/values?tag=value`).then(
      (response) => {
        expect(response.body).to.have.property("timestamp").that.is.a("string")
        expect(Date.parse(response.body.timestamp)).to.not.be.NaN
        expect(response.body).to.have.deep.property("tags", { tag: "value" })
        expect(response.body).to.have.deep.property("fields", {
          2994: false,
          2263: "open62541",
          "the.answer": 42,
          myByteString: "dGVzdDEyMw==",
        })
      }
    )
  })
})
