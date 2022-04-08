#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@confidential-storage-hub
Feature: Confidential Storage Hub
  Background: Components are ready
    # confidential storage hub
    When an HTTP GET is sent to "https://localhost:8095/healthcheck"
    Then response status is "200 OK"
     And response contains "status" with value "success"
    # hub-kms
    When an HTTP GET is sent to "https://localhost:8074/healthcheck"
    Then response status is "200 OK"
     And response contains "status" with value "success"
    # edv
    When an HTTP GET is sent to "https://localhost:8081/healthcheck"
    Then response status is "200 OK"
     And response contains "status" with value "success"

  Scenario: Create profile
    When the user requests a new confidential-storage-hub profile
    Then the confidential-storage-hub profile is created

  Scenario: Comparison between two equal documents with doc queries
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hello World!"
     And the user saves a Confidential Storage document with content "Hello World!"
     And the user authorizes the CSH to read the documents
     And the user requests a comparison between the two documents
    Then the result is "true"

  Scenario: Comparison between two unequal documents with doc queries
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hey Alice!"
     And the user saves a Confidential Storage document with content "Goodbye Bob!"
     And the user authorizes the CSH to read the documents
     And the user requests a comparison between the two documents
    Then the result is "false"

  Scenario: Comparison between two equal documents - 1 DocQuery, 1 RefQuery
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hello World!"
     And the user saves a Confidential Storage document with content "Hello World!"
     And the user authorizes the CSH to read the documents
     And the user creates a RefQuery for one document
     And the user requests a comparison between the two documents
    Then the result is "true"

  Scenario: Comparison between two equal documents - 2 RefQuery
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hello World!"
     And the user saves a Confidential Storage document with content "Hello World!"
     And the user authorizes the CSH to read the documents
     And the user creates a RefQuery for one document
     And the user creates a RefQuery for one document
     And the user requests a comparison between the two documents
    Then the result is "true"

  Scenario: Comparison between two unequal documents - 1 DocQuery, 1 RefQuery
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hey Alice!"
     And the user saves a Confidential Storage document with content "Goodbye Bob!"
     And the user authorizes the CSH to read the documents
     And the user creates a RefQuery for one document
     And the user requests a comparison between the two documents
    Then the result is "false"

  Scenario: Comparison between two unequal documents - 2 RefQuery
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hey Alice!"
     And the user saves a Confidential Storage document with content "Goodbye Bob!"
     And the user authorizes the CSH to read the documents
     And the user creates a RefQuery for one document
     And the user creates a RefQuery for one document
     And the user requests a comparison between the two documents
    Then the result is "false"

  Scenario: Extraction of several documents with DocQuery and RefQuery
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hey Alice!"
     And the user saves a Confidential Storage document with content "Goodbye Bob!"
     And the user authorizes the CSH to read the documents
     And the user creates a RefQuery for one document
     And the user requests extraction of all documents
    Then the CSH returns the decrypted documents

  Scenario: Extraction of several documents with RefQuery
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hey Alice!"
     And the user saves a Confidential Storage document with content "Goodbye Bob!"
     And the user authorizes the CSH to read the documents
     And the user creates a RefQuery for one document
     And the user creates a RefQuery for one document
     And the user requests extraction of all documents
    Then the CSH returns the decrypted documents

  Scenario: Extraction of several documents with DocQuery
    When the user has a profile
     And the user saves a Confidential Storage document with content "Hey Alice!"
     And the user saves a Confidential Storage document with content "Goodbye Bob!"
     And the user authorizes the CSH to read the documents
     And the user requests extraction of all documents
    Then the CSH returns the decrypted documents
