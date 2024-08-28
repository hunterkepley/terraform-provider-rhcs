/*
Copyright (c) 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package classic

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

const managedOidcConfig = `{
  "href": "/api/clusters_mgmt/v1/oidc_configs/23f6gk51qi5ng15mm095c90hhajbf7c5",
  "id": "23f6gk51qi5ng15mm095c90hhajbf7c5",
  "issuer_url": "https://d3gt1gce2zmg3d.cloudfront.net/23f6gk51qi5ng15mm095c90hhajbf7c5",
  "managed": true,
  "reusable": true
}`

const oidcConfigThumbprint = `{
  "href": "/api/clusters_mgmt/v1/aws_inquiries/oidc_thumbprint/9e99a48a9960b14926bb7f3b02e22da2b0ab7280",
  "thumbprint": "9e99a48a9960b14926bb7f3b02e22da2b0ab7280",
  "oidc_config_id": "23f6gk51qi5ng15mm095c90hhajbf7c5",
  "cluster_id": ""
}`

const unManagedOidcConfig = `{
  "href": "/api/clusters_mgmt/v1/oidc_configs/23f6gk51qi5ng15mm095c90hhajbf7c5",
  "id": "23f6gk51qi5ng15mm095c90hhajbf7c5",
  "issuer_url": "https://oidc-f3y4.s3.us-east-1.amazonaws.com",
  "secret_arn": "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c",
  "managed": false,
  "reusable": true
}`

const clusterListIsEmpty = `{
  "kind": "ClusterList",
  "page": 0,
  "size": 0,
  "total": 0,
  "items": [
  ]
}`
const clusterListIsNotEmpty = `{
  "kind": "ClusterList",
  "page": 1,
  "size": 1,
  "total": 1,
  "items": [
		{
			"name": "cluster-name"
		}
  ]
}`

const getOidcConfigURL = "/api/clusters_mgmt/v1/oidc_configs/23f6gk51qi5ng15mm095c90hhajbf7c5"
const getOidcConfigThumbprintURL = "/api/clusters_mgmt/v1/aws_inquiries/oidc_thumbprint"
const installerRoleARN = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
const unManagedIssuerURL = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
const managedIssuerURL = "https://d3gt1gce2zmg3d.cloudfront.net/23f6gk51qi5ng15mm095c90hhajbf7c5"
const managedOidcEndpointURL = "d3gt1gce2zmg3d.cloudfront.net/23f6gk51qi5ng15mm095c90hhajbf7c5"
const unManagedOidcEndpointURL = "oidc-f3y4.s3.us-east-1.amazonaws.com"
const secretARN = "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
const ID = "23f6gk51qi5ng15mm095c90hhajbf7c5"
const thumbprint = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"

var _ = Describe("OIDC config creation", func() {
	It("Can create managed OIDC config", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/oidc_configs"),
				VerifyJQ(`.managed`, true),
				RespondWithJSON(http.StatusOK, managedOidcConfig),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, getOidcConfigThumbprintURL),
				RespondWithJSON(http.StatusCreated, oidcConfigThumbprint),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL),
				RespondWithJSON(http.StatusOK, managedOidcConfig),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, getOidcConfigThumbprintURL),
				RespondWithJSON(http.StatusCreated, oidcConfigThumbprint),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL),
				RespondWithJSON(http.StatusOK, managedOidcConfig),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"),
				RespondWithJSON(http.StatusOK, clusterListIsEmpty),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, getOidcConfigURL),
				RespondWithJSON(http.StatusNoContent, managedOidcConfig),
			),
		)

		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = true
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
		resource := Terraform.Resource("rhcs_rosa_oidc_config", "oidc_config")
		Expect(resource).To(MatchJQ(".attributes.id", ID))
		Expect(resource).To(MatchJQ(".attributes.issuer_url", managedIssuerURL))
		Expect(resource).To(MatchJQ(".attributes.managed", true))
		Expect(resource).To(MatchJQ(".attributes.thumbprint", thumbprint))
		Expect(resource).To(MatchJQ(".attributes.oidc_endpoint_url", managedOidcEndpointURL))
		Expect(Terraform.Destroy().ExitCode).To(BeZero())
	})

	Context("Create unmanaged OIDC config", func() {
		BeforeEach(func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/oidc_configs"),
					VerifyJQ(`.managed`, false),
					VerifyJQ(`.installer_role_arn`, installerRoleARN),
					VerifyJQ(`.issuer_url`, unManagedIssuerURL),
					VerifyJQ(`.secret_arn`, secretARN),
					RespondWithJSON(http.StatusOK, unManagedOidcConfig),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, getOidcConfigThumbprintURL),
					RespondWithJSON(http.StatusCreated, oidcConfigThumbprint),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, getOidcConfigURL),
					RespondWithJSON(http.StatusOK, unManagedOidcConfig),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, getOidcConfigThumbprintURL),
					RespondWithJSON(http.StatusCreated, oidcConfigThumbprint),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, getOidcConfigURL),
					RespondWithJSON(http.StatusOK, unManagedOidcConfig),
				),
			)
		})
		It("Succeed to destroy it", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"),
					RespondWithJSON(http.StatusOK, clusterListIsEmpty),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL),
					RespondWithJSON(http.StatusNoContent, unManagedOidcConfig),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
		}
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			validateTerraformResourceState()
			Expect(Terraform.Destroy().ExitCode).To(BeZero())
		})

		It("Fail on destroy due to a cluster that using it", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"),
					RespondWithJSON(http.StatusOK, clusterListIsNotEmpty),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL),
					RespondWithJSON(http.StatusNoContent, unManagedOidcConfig),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
		}
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			validateTerraformResourceState()

			// fail on destroy
			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("there are clusters using OIDC config")

		})

		It("Fail on destroy because fail to get if there is a cluster that using it", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"),
					RespondWithJSON(http.StatusNotFound, clusterListIsNotEmpty),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL),
					RespondWithJSON(http.StatusNoContent, unManagedOidcConfig),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
		}
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			validateTerraformResourceState()

			// fail on destroy
			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("There was a problem checking if any clusters are using OIDC config")
		})

		It("Fail on destroy because fail to remove the oidc config resource from OCM", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"),
					RespondWithJSON(http.StatusOK, clusterListIsEmpty),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL),
					RespondWithJSON(http.StatusInternalServerError, unManagedOidcConfig),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
		}
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			validateTerraformResourceState()

			// fail on destroy
			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("There was a problem deleting the OIDC config")
		})
	})

	It("Try to create managed OIDC config with unsupported attributes and fail", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/oidc_configs"),
				VerifyJQ(`.managed`, true),
				VerifyJQ(`.installer_role_arn`, installerRoleARN),
				VerifyJQ(`.issuer_url`, unManagedIssuerURL),
				VerifyJQ(`.secret_arn`, secretARN),
				RespondWithJSON(http.StatusOK, unManagedOidcConfig),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL),
				RespondWithJSON(http.StatusOK, unManagedOidcConfig),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL),
				RespondWithJSON(http.StatusOK, unManagedOidcConfig),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"),
				RespondWithJSON(http.StatusOK, clusterListIsEmpty),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, getOidcConfigURL),
				RespondWithJSON(http.StatusNoContent, unManagedOidcConfig),
			),
		)

		// Run the apply command:
		Terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = true
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
		}
		`)
		// expect to fail
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("In order to create managed OIDC Configuration, the attributes' values of `secret_arn`, `issuer_url` and `installer_role_arn` should be empty")
	})

})

func validateTerraformResourceState() {
	resource := Terraform.Resource("rhcs_rosa_oidc_config", "oidc_config")
	Expect(resource).To(MatchJQ(".attributes.id", ID))
	Expect(resource).To(MatchJQ(".attributes.installer_role_arn", installerRoleARN))
	Expect(resource).To(MatchJQ(".attributes.managed", false))
	Expect(resource).To(MatchJQ(".attributes.issuer_url", unManagedIssuerURL))
	Expect(resource).To(MatchJQ(".attributes.secret_arn", secretARN))
	Expect(resource).To(MatchJQ(".attributes.thumbprint", thumbprint))
	Expect(resource).To(MatchJQ(".attributes.oidc_endpoint_url", unManagedOidcEndpointURL))

}
