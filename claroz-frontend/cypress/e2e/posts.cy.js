describe('Post Interactions', () => {
  beforeEach(() => {
    // Login before each test
    cy.login();
    cy.visit('/');
  });

  it('should create a new post', () => {
    const postContent = 'This is a test post ' + Date.now();
    
    // Create post
    cy.createPost(postContent);
    
    // Verify post appears in feed
    cy.get('.MuiCard-root')
      .first()
      .within(() => {
        cy.contains(postContent).should('be.visible');
        cy.contains('testuser').should('be.visible');
        cy.get('button[aria-label="like"]').should('exist');
        cy.get('button[aria-label="comment"]').should('exist');
      });
  });

  it('should like and unlike a post', () => {
    const postContent = 'Post for like testing ' + Date.now();
    
    // Create post
    cy.createPost(postContent);
    
    // Find the post and verify initial like state
    cy.contains(postContent)
      .closest('.MuiCard-root')
      .within(() => {
        // Initial state - not liked
        cy.get('button[aria-label="like"]')
          .should('not.have.class', 'Mui-active');
        
        // Like the post
        cy.get('button[aria-label="like"]').click();
        
        // Verify liked state
        cy.get('button[aria-label="like"]')
          .should('have.class', 'Mui-active');
        
        // Unlike the post
        cy.get('button[aria-label="like"]').click();
        
        // Verify returned to unliked state
        cy.get('button[aria-label="like"]')
          .should('not.have.class', 'Mui-active');
      });
  });

  it('should add and view comments', () => {
    const postContent = 'Post for comment testing ' + Date.now();
    const comment = 'This is a test comment ' + Date.now();
    
    // Create post
    cy.createPost(postContent);
    
    // Add comment
    cy.addComment(postContent, comment);
    
    // Verify comment appears
    cy.contains(postContent)
      .closest('.MuiCard-root')
      .within(() => {
        cy.contains(comment).should('be.visible');
        cy.contains('testuser').should('be.visible');
      });
  });

  it('should handle long posts', () => {
    const longPost = 'A'.repeat(500) + Date.now();
    
    // Create long post
    cy.createPost(longPost);
    
    // Verify post appears correctly
    cy.contains(longPost).should('be.visible');
  });

  it('should validate empty posts', () => {
    // Try to submit empty post
    cy.get('button[type="submit"]').should('be.disabled');
    
    // Try to submit whitespace
    cy.get('textarea[name="content"]').type('   ');
    cy.get('button[type="submit"]').should('be.disabled');
  });

  it('should handle post creation errors', () => {
    // Force network error
    cy.intercept('POST', '/api/posts', {
      statusCode: 500,
      body: { message: 'Server error' }
    });
    
    const postContent = 'Error test post ' + Date.now();
    cy.createPost(postContent);
    
    // Verify error message
    cy.get('.MuiAlert-root')
      .should('be.visible')
      .and('contain', 'Failed to create post');
  });

  it('should load more posts on scroll', () => {
    // Get initial post count
    cy.get('.MuiCard-root').then(($posts) => {
      const initialCount = $posts.length;
      
      // Scroll to bottom
      cy.get('.MuiCard-root').last().scrollIntoView();
      
      // Verify more posts loaded
      cy.get('.MuiCard-root').should('have.length.greaterThan', initialCount);
    });
  });

  it('should handle multiple comments on a post', () => {
    const postContent = 'Multi-comment test post ' + Date.now();
    const comments = [
      'First comment ' + Date.now(),
      'Second comment ' + Date.now(),
      'Third comment ' + Date.now()
    ];
    
    // Create post
    cy.createPost(postContent);
    
    // Add multiple comments
    comments.forEach(comment => {
      cy.addComment(postContent, comment);
    });
    
    // Verify all comments appear
    cy.contains(postContent)
      .closest('.MuiCard-root')
      .within(() => {
        comments.forEach(comment => {
          cy.contains(comment).should('be.visible');
        });
      });
  });

  it('should handle post deletion', () => {
    const postContent = 'Post to delete ' + Date.now();
    
    // Create post
    cy.createPost(postContent);
    
    // Delete post
    cy.contains(postContent)
      .closest('.MuiCard-root')
      .within(() => {
        cy.get('[aria-label="delete"]').click();
      });
    
    // Confirm deletion
    cy.get('[role="dialog"]')
      .within(() => {
        cy.contains('Delete').click();
      });
    
    // Verify post is removed
    cy.contains(postContent).should('not.exist');
  });
});
