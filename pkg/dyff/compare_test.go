// Copyright © 2018 Matthias Diester
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package dyff_test

import (
	. "github.com/HeavyWombat/dyff/pkg/dyff"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Core/Compare", func() {
	Describe("Difference between YAMLs", func() {
		Context("Given two simple YAML structures", func() {
			It("should return that a string value was modified", func() {
				from := yml(`---
some:
  yaml:
    structure:
      name: foobar
      version: v1
`)

				to := yml(`---
some:
  yaml:
    structure:
      name: fOObAr
      version: v1
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(len(result[0].Details)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/name", MODIFICATION, "foobar", "fOObAr")))
			})

			It("should return that an integer was modified", func() {
				from := yml(`---
some:
  yaml:
    structure:
      name: 1
      version: v1
`)

				to := yml(`---
some:
  yaml:
    structure:
      name: 2
      version: v1
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/name", MODIFICATION, 1, 2)))
			})

			It("should return that a float was modified", func() {
				from := yml(`---
some:
  yaml:
    structure:
      name: foobar
      level: 3.14159265359
`)

				to := yml(`---
some:
  yaml:
    structure:
      name: foobar
      level: 2.7182818284
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/level", MODIFICATION, 3.14159265359, 2.7182818284)))
			})

			It("should return that a boolean was modified", func() {
				from := yml(`---
some:
  yaml:
    structure:
      name: foobar
      enabled: false
`)

				to := yml(`---
some:
  yaml:
    structure:
      name: foobar
      enabled: true
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/enabled", MODIFICATION, false, true)))
			})

			It("should return that one value was added", func() {
				from := yml(`---
some:
  yaml:
    structure:
      name: foobar
`)

				to := yml(`---
some:
  yaml:
    structure:
      name: foobar
      version: v1
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure", ADDITION, nil, yml(`version: v1`))))
			})

			It("should return that one value was removed", func() {
				from := yml(`---
some:
  yaml:
    structure:
      name: foobar
      version: v1
`)

				to := yml(`---
some:
  yaml:
    structure:
      name: foobar
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure", REMOVAL, yml(`version: v1`), nil)))
			})

			It("should return two diffs if one value was removed and another one added", func() {
				from := yml(`---
some:
  yaml:
    structure:
      name: foobar
      version: v1
`)

				to := yml(`---
some:
  yaml:
    structure:
      name: foobar
      release: v1
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(doubleDiff("/some/yaml/structure",
					REMOVAL, yml(`version: v1`), nil,
					ADDITION, nil, yml(`release: v1`))))
			})
		})

		Context("Given two YAML structures with simple lists", func() {
			It("should return that a string list entry was added", func() {
				from := yml(`---
some:
  yaml:
    structure:
      list:
      - one
      - two
`)

				to := yml(`---
some:
  yaml:
    structure:
      list:
      - one
      - two
      - three
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/list", ADDITION, nil, yml(`list: [ three ]`)[0].Value)))
			})

			It("should return that an integer list entry was added", func() {
				from := yml(`---
some:
  yaml:
    structure:
      list:
      - 1
      - 2
`)

				to := yml(`---
some:
  yaml:
    structure:
      list:
      - 1
      - 2
      - 3
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))

				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/list", ADDITION, nil, yml(`list: [ 3 ]`)[0].Value)))
			})

			It("should return that a string list entry was removed", func() {
				from := yml(`---
some:
  yaml:
    structure:
      list:
      - one
      - two
      - three
`)

				to := yml(`---
some:
  yaml:
    structure:
      list:
      - one
      - two
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/list", REMOVAL, yml(`list: [ three ]`)[0].Value, nil)))
			})

			It("should return that an integer list entry was removed", func() {
				from := yml(`---
some:
  yaml:
    structure:
      list:
      - 1
      - 2
      - 3
`)

				to := yml(`---
some:
  yaml:
    structure:
      list:
      - 1
      - 2
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/some/yaml/structure/list", REMOVAL, yml(`list: [ 3 ]`)[0].Value, nil)))
			})

			It("should not return a change if only the order in a hash was changed", func() {
				from := yml(`---
list:
- enabled: true
- foo: bar
  version: 1
`)

				to := yml(`---
list:
- enabled: true
- version: 1
  foo: bar
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(0))
			})
		})

		Context("Given two YAML structures with complex content", func() {
			It("should return all differences in there", func() {
				from := yml(`---
instance_groups:
- name: web
  instances: 1
  resource_pool: concourse_resource_pool
  networks:
  - name: concourse
    static_ips: 192.168.1.1
  jobs:
  - release: concourse
    name: atc
    properties:
      postgresql_database: &atc-db atc
      external_url: http://192.168.1.100:8080
      development_mode: true
  - release: concourse
    name: tsa
    properties: {}

  - name: db
    instances: 1
    resource_pool: concourse_resource_pool
    networks: [{name: concourse}, {name: testnet}]
    persistent_disk: 10240
    jobs:
    - release: concourse
      name: postgresql
      properties:
        databases:
        - name: *atc-db
          role: atc
          password: supersecret
`)

				to := yml(`---
instance_groups:
- name: web
  instances: 1
  resource_pool: concourse_resource_pool
  networks:
  - name: concourse
    static_ips: 192.168.0.1
  jobs:
  - release: concourse
    name: atc
    properties:
      postgresql_database: &atc-db atc
      external_url: http://192.168.0.100:8080
      development_mode: false
  - release: concourse
    name: tsa
    properties: {}
  - release: custom
    name: logger

  - name: db
    instances: 2
    resource_pool: concourse_resource_pool
    networks: [{name: concourse}]
    persistent_disk: 10240
    jobs:
    - release: concourse
      name: postgresql
      properties:
        databases:
        - name: *atc-db
          role: atc
          password: "zwX#(;P=%hTfFzM["
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(7))

				Expect(result[0]).To(BeEquivalentTo(singleDiff("/instance_groups/name=web/networks/name=concourse/static_ips", MODIFICATION, "192.168.1.1", "192.168.0.1")))
				Expect(result[1]).To(BeEquivalentTo(singleDiff("/instance_groups/name=web/jobs", ADDITION, nil, yml(`list: [ { release: custom, name: logger } ]`)[0].Value)))
				Expect(result[2]).To(BeEquivalentTo(singleDiff("/instance_groups/name=web/jobs/name=atc/properties/external_url", MODIFICATION, "http://192.168.1.100:8080", "http://192.168.0.100:8080")))
				Expect(result[3]).To(BeEquivalentTo(singleDiff("/instance_groups/name=web/jobs/name=atc/properties/development_mode", MODIFICATION, true, false)))
				Expect(result[4]).To(BeEquivalentTo(singleDiff("/instance_groups/name=web/jobs/name=db/instances", MODIFICATION, 1, 2)))
				Expect(result[5]).To(BeEquivalentTo(singleDiff("/instance_groups/name=web/jobs/name=db/networks", REMOVAL, yml(`list: [ { name: testnet } ]`)[0].Value, nil)))
				Expect(result[6]).To(BeEquivalentTo(singleDiff("/instance_groups/name=web/jobs/name=db/jobs/name=postgresql/properties/databases/name=atc/password", MODIFICATION, "supersecret", "zwX#(;P=%hTfFzM[")))
			})

			It("should return even difficult ones", func() {
				from := yml(`---
resource_pools:
- name: concourse_resource_pool
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: '3232.2'
  network: concourse
  cloud_properties:
    ram: 4096
    disk: 32768
    cpu: 2
    datacenters:
    - clusters:
      - CLS_PAAS_SFT_035: {resource_pool: other-vsphere-res-pool}
`)

				to := yml(`---
resource_pools:
- name: concourse_resource_pool
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: '3232.2'
  network: concourse
  cloud_properties:
    ram: 4096
    disk: 32768
    cpu: 2
    datacenters:
    - clusters:
      - CLS_PAAS_SFT_035:
          resource_pool: new-vsphere-res-pool
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/resource_pools/name=concourse_resource_pool/cloud_properties/datacenters/0/clusters/0/CLS_PAAS_SFT_035/resource_pool", MODIFICATION, "other-vsphere-res-pool", "new-vsphere-res-pool")))
			})

			It("should return even difficult ones that cannot simply be compared", func() {
				from := yml(`---
resource_pools:
- name: concourse_resource_pool
  cloud_properties:
    datacenters:
    - clusters:
      - CLS_PAAS_SFT_035: {resource_pool: 35-vsphere-res-pool}
      - CLS_PAAS_SFT_036: {resource_pool: 36-vsphere-res-pool}
`)

				to := yml(`---
resource_pools:
- name: concourse_resource_pool
  cloud_properties:
    datacenters:
    - clusters:
      - CLS_PAAS_SFT_035: {resource_pool: 35a-vsphere-res-pool}
      - CLS_PAAS_SFT_036: {resource_pool: 36a-vsphere-res-pool}
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(doubleDiff("/resource_pools/name=concourse_resource_pool/cloud_properties/datacenters/0/clusters",
					REMOVAL, yml(`list: [ {CLS_PAAS_SFT_035: {resource_pool: 35-vsphere-res-pool}}, {CLS_PAAS_SFT_036: {resource_pool: 36-vsphere-res-pool}} ]`)[0].Value, nil,
					ADDITION, nil, yml(`list: [ {CLS_PAAS_SFT_035: {resource_pool: 35a-vsphere-res-pool}}, {CLS_PAAS_SFT_036: {resource_pool: 36a-vsphere-res-pool}} ]`)[0].Value)))
			})
		})

		Context("Given two YAMLs with a list as the root", func() {
			It("should return the differences the same way", func() {
				from := cmplxList(`---
- name: one
  version: 1

- name: two
  version: 2

- name: three
  version: 4
`)

				to := cmplxList(`---
- name: one
  version: 1

- name: two
  version: 2

- name: three
  version: 3
`)

				result := CompareDocuments(from, to)
				Expect(result).NotTo(BeNil())
				Expect(len(result)).To(BeEquivalentTo(1))
				Expect(result[0]).To(BeEquivalentTo(singleDiff("/name=three/version",
					MODIFICATION, 4, 3)))
			})
		})

		Context("Given two YAML files", func() {
			It("should return all differences between the files", func() {
				results := CompareDocuments(yml("../../assets/examples/from.yml"), yml("../../assets/examples/to.yml"))
				expected := []Diff{
					doubleDiff("/yaml/map",
						REMOVAL, yml(`---
stringB: fOObAr
intB: 10
floatB: 2.71
boolB: false
mapB: { key0: B, key1: B }
listB: [ B, B, B ]
`), nil,
						ADDITION, nil, yml(`---
stringY: YAML!
intY: 147
floatY: 24.0
boolY: true
mapY: { key0: Y, key1: Y }
listY: [ Yo, Yo, Yo ]
`)),

					singleDiff("/yaml/map/type-change-1", MODIFICATION, "string", 147),

					singleDiff("/yaml/map/type-change-2", MODIFICATION, "12", 12),

					singleDiff("/yaml/map/whitespaces", MODIFICATION, "Strings can  have whitespaces.", "Strings can  have whitespaces.\n\n\n"),

					doubleDiff("/yaml/simple-list",
						REMOVAL, yml(`list: [ X, Z ]`)[0].Value, nil,
						ADDITION, nil, yml(`list: [ D, E ]`)[0].Value),

					doubleDiff("/yaml/named-entry-list-using-name",
						REMOVAL, yml(`list: [ {name: X}, {name: Z} ]`)[0].Value, nil,
						ADDITION, nil, yml(`list: [ {name: D}, {name: E} ]`)[0].Value),

					doubleDiff("/yaml/named-entry-list-using-key",
						REMOVAL, yml(`list: [ {key: X}, {key: Z} ]`)[0].Value, nil,
						ADDITION, nil, yml(`list: [ {key: D}, {key: E} ]`)[0].Value),

					doubleDiff("/yaml/named-entry-list-using-id",
						REMOVAL, yml(`list: [ {id: X}, {id: Z} ]`)[0].Value, nil,
						ADDITION, nil, yml(`list: [ {id: D}, {id: E} ]`)[0].Value),
				}

				Expect(results).NotTo(BeNil())
				Expect(len(results)).To(BeEquivalentTo(len(expected)))

				for i, result := range results {
					Expect(result).To(BeEquivalentTo(expected[i]))
				}
			})

			It("should return order changes in named entry lists (ignoring additions and removals)", func() {
				from := yml(`list: [ {name: A}, {name: C}, {name: B}, {name: D}, {name: E} ]`)
				to := yml(`list: [ {name: A}, {name: X1}, {name: B}, {name: C}, {name: D}, {name: X2} ]`)
				results := CompareDocuments(from, to)

				Expect(results).NotTo(BeNil())
				Expect(len(results)).To(BeEquivalentTo(1))
				Expect(len(results[0].Details)).To(BeEquivalentTo(3))
				Expect(results[0].Details[0]).To(BeEquivalentTo(Detail{
					Kind: ORDERCHANGE,
					From: []string{"A", "C", "B", "D"},
					To:   []string{"A", "B", "C", "D"},
				}))
			})

			It("should return order changes in simple lists (ignoring additions and removals)", func() {
				from := yml(`list: [ A, C, B, D, E ]`)
				to := yml(`list: [ A, X1, B, C, D, X2 ]`)
				results := CompareDocuments(from, to)

				Expect(results).NotTo(BeNil())
				Expect(len(results)).To(BeEquivalentTo(1))
				Expect(len(results[0].Details)).To(BeEquivalentTo(3))
				Expect(results[0].Details[0]).To(BeEquivalentTo(Detail{
					Kind: ORDERCHANGE,
					From: []interface{}{"A", "C", "B", "D"},
					To:   []interface{}{"A", "B", "C", "D"},
				}))
			})

			It("should return all differences between the files with multipe documents", func() {
				results := CompareInputFiles(file("../../assets/kubernetes-yaml/from.yml"), file("../../assets/kubernetes-yaml/to.yml"))
				expected := []Diff{

					singleDiff("#0/spec/template/spec/containers/name=registry/resources/limits/cpu", MODIFICATION, "100m", "1000m"),
					singleDiff("#0/spec/template/spec/containers/name=registry/resources/limits/memory", MODIFICATION, "100Mi", "10Gi"),
					singleDiff("#0/spec/template/spec/containers/name=registry/ports", ADDITION, nil, yml(`list: [ {containerPort: 5001, name: backdoor, protocol: TCP} ]`)[0].Value),
					singleDiff("#1/spec/ports", ADDITION, nil, yml(`list: [ {name: backdoor, port: 5001, protocol: TCP} ]`)[0].Value),
				}

				Expect(results).NotTo(BeNil())
				Expect(len(results)).To(BeEquivalentTo(len(expected)))

				for i, result := range results {
					Expect(result).To(BeEquivalentTo(expected[i]))
				}
			})
		})
	})
})
