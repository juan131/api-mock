describe('API mock', () => {
  it('Heartbeat & metrics endpoints', () => {
    cy.request('/live').as('getLive');
    cy.get('@getLive').then(res => {
      expect(res.status).to.eq(200);
    });
    cy.request('/ready').as('getReady');
    cy.get('@getReady').then(res => {
      expect(res.status).to.eq(200);
    });
  });

  it('API endpoints (authn)', () => {
    cy.request({
      method: 'GET',
      url: '/v1/mock/foo',
      failOnStatusCode: false,
    }).as('noAuthn');
    cy.get('@noAuthn').then(res => {
      expect(res.status).to.eq(401);
    });
    cy.request({
      method: 'GET',
      url: '/v1/mock/foo',
      headers: {'X-API-KEY': 'some-api-key'},
      failOnStatusCode: false,
    }).as('getFoo');
    cy.get('@getFoo').then(res => {
      expect(res.status).to.eq(200);
      expect(res.body).to.have.property('message');
      expect(res.body.message).to.eq('success');
    });
  });

  it('API endpoints (success ratio)', () => {
    cy.request({
      method: 'GET',
      url: '/v1/mock/foo',
      headers: {'X-API-KEY': 'some-api-key'},
      failOnStatusCode: false,
    }).as('resetCounter');
    cy.get('@resetCounter').then(res => {
      cy.request({
        method: 'GET',
        url: '/v1/mock/bar',
        headers: {'X-API-KEY': 'some-api-key'},
        failOnStatusCode: false,
      }).as('getBar');
      cy.get('@getBar').then(res => {
        expect(res.status).to.eq(200);
        expect(res.body).to.have.property('message');
        expect(res.body.message).to.eq('success');
        cy.request({
          method: 'POST',
          url: '/v1/mock/bar',
          headers: {'X-API-KEY': 'some-api-key'},
          body: JSON.stringify({foo: 'bar'}),
          failOnStatusCode: false,
        }).as('postBar');
        cy.get('@postBar').then(res => {
          expect(res.status).to.eq(400);
          expect(res.body).to.have.property('error');
          expect(res.body.error).to.have.property('code');
          expect(res.body.error).to.have.property('message');
          expect(res.body.error.code).to.eq(1005);
          expect(res.body.error.message).to.eq('failed request');
        });
      });
    });
  });
});
