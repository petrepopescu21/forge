Feature: Health Check
  As a user
  I want the API to provide a health check endpoint
  So that I can verify the service is running

  Scenario: API should return healthy status
    Given the API server is running
    When I make a GET request to "/api/v1/health"
    Then the response status should be 200
    And the response body should contain "healthy": true
