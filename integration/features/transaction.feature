Feature: transaction
  In order to pay things
  As a user
  I need to be able to create transactions

  Scenario: Create a transaction
    Given I have no transactions
    When I create a new transactions
    Then there should be 1 transaction