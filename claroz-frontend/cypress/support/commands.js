// Custom commands for Cypress tests

Cypress.Commands.add('login', (email = 'test@example.com', password = 'password123') => {
  cy.visit('/login');
  cy.get('input[name="email"]').type(email);
  cy.get('input[name="password"]').type(password);
  cy.get('button[type="submit"]').click();
  // Wait for navigation to complete
  cy.url().should('eq', Cypress.config().baseUrl + '/');
});

// Command to create a new post
Cypress.Commands.add('createPost', (content) => {
  cy.get('textarea[name="content"]').type(content);
  cy.get('button[type="submit"]').click();
  // Wait for post to appear
  cy.contains(content).should('be.visible');
});

// Command to like a post
Cypress.Commands.add('likePost', (postContent) => {
  cy.contains(postContent)
    .closest('.MuiCard-root')
    .within(() => {
      cy.get('button[aria-label="like"]').click();
    });
});

// Command to add a comment to a post
Cypress.Commands.add('addComment', (postContent, comment) => {
  cy.contains(postContent)
    .closest('.MuiCard-root')
    .within(() => {
      cy.get('button[aria-label="comment"]').click();
      cy.get('input[placeholder="Add a comment..."]').type(comment);
      cy.get('button[type="submit"]').click();
    });
  // Wait for comment to appear
  cy.contains(comment).should('be.visible');
});
