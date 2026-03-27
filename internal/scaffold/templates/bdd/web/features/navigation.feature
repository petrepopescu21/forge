Feature: Navigation
  As a user
  I want to navigate between pages
  So that I can access different parts of the application

  Scenario: User can navigate to the leads page
    Given I am on the home page
    When I click the "Leads" navigation link
    Then I should be on the "/leads" page
    And the page title should contain "Leads"

  Scenario: User can navigate to the dashboard
    Given I am on the home page
    When I click the "Dashboard" navigation link
    Then I should be on the "/dashboard" page
