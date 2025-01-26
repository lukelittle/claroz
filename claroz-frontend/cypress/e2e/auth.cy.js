describe('Authentication', () => {
  beforeEach(() => {
    // Reset any previous state
    cy.window().then((win) => {
      win.localStorage.clear();
    });
  });

  it('should successfully log in with valid credentials', () => {
    cy.login();
    // Verify successful login
    cy.url().should('eq', Cypress.config().baseUrl + '/');
    cy.get('[aria-label="user menu"]').should('exist');
  });

  it('should show error message with invalid credentials', () => {
    cy.visit('/login');
    cy.get('input[name="email"]').type('wrong@example.com');
    cy.get('input[name="password"]').type('wrongpassword');
    cy.get('button[type="submit"]').click();
    
    // Verify error message
    cy.get('.MuiAlert-root').should('be.visible')
      .and('contain', 'Invalid credentials');
    
    // Verify we're still on login page
    cy.url().should('include', '/login');
  });

  it('should navigate to register page and back', () => {
    cy.visit('/login');
    cy.contains("Don't have an account? Sign Up").click();
    cy.url().should('include', '/register');
    
    // Verify register form elements
    cy.get('input[name="username"]').should('exist');
    cy.get('input[name="email"]').should('exist');
    cy.get('input[name="password"]').should('exist');
    cy.get('input[name="confirmPassword"]').should('exist');
    
    // Navigate back to login
    cy.contains('Already have an account? Sign In').click();
    cy.url().should('include', '/login');
  });

  it('should require all fields for login', () => {
    cy.visit('/login');
    
    // Try submitting without any fields
    cy.get('button[type="submit"]').click();
    cy.get('input[name="email"]:invalid').should('exist');
    cy.get('input[name="password"]:invalid').should('exist');
    
    // Try submitting with only email
    cy.get('input[name="email"]').type('test@example.com');
    cy.get('button[type="submit"]').click();
    cy.get('input[name="password"]:invalid').should('exist');
    
    // Try submitting with only password
    cy.get('input[name="email"]').clear();
    cy.get('input[name="password"]').type('password123');
    cy.get('button[type="submit"]').click();
    cy.get('input[name="email"]:invalid').should('exist');
  });

  it('should persist login state after page reload', () => {
    cy.login();
    cy.reload();
    
    // Verify we're still logged in
    cy.url().should('eq', Cypress.config().baseUrl + '/');
    cy.get('[aria-label="user menu"]').should('exist');
  });

  it('should successfully log out', () => {
    cy.login();
    
    // Click user menu and logout
    cy.get('[aria-label="user menu"]').click();
    cy.contains('Logout').click();
    
    // Verify logout
    cy.url().should('include', '/login');
    cy.window().its('localStorage').should('be.empty');
  });

  it('should redirect to login when accessing protected routes while logged out', () => {
    cy.visit('/profile');
    cy.url().should('include', '/login');
    
    cy.visit('/feed');
    cy.url().should('include', '/login');
  });
});
