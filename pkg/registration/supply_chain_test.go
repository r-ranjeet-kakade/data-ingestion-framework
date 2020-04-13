package registration

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"gopkg.in/yaml.v2"
	"log"
	"testing"
)

func TestSupplyChainNode(t *testing.T) {
	filename := "test_app_node.yaml"

	var supplyChainConfig *conf.SupplyChainConfig
	supplyChainConfig, err := conf.LoadSupplyChain(filename)
	if err != nil {
		fmt.Printf("%++v\n", err)
	}
	assert.True(t, err == nil)
	assert.True(t, len(supplyChainConfig.Nodes) > 0)

	supplyChain, err := NewSupplyChain(supplyChainConfig)
	if err != nil {
		fmt.Printf("%++v\n", err)
	}
	assert.True(t, err == nil)

	entityType := proto.EntityDTO_APPLICATION_COMPONENT
	nodes := supplyChain.nodeMap

	if _, exists := nodes[entityType]; !exists {
		t.Errorf("Missing %s node", entityType)
		assert.Fail(t, "Missing %s node", entityType)
	}

	appNode := nodes[entityType]

	expectedSoldComms := []proto.CommodityDTO_CommodityType{
		proto.CommodityDTO_TRANSACTION,
		proto.CommodityDTO_RESPONSE_TIME,
		proto.CommodityDTO_COLLECTION_TIME,
		proto.CommodityDTO_THREADS,
		proto.CommodityDTO_HEAP,
	}

	var soldCommsList []proto.CommodityDTO_CommodityType
	for key, _ := range appNode.SupportedComms {
		soldCommsList = append(soldCommsList, key)
	}

	assert.ElementsMatch(t, expectedSoldComms, soldCommsList)

	expectedProviders := []proto.EntityDTO_EntityType{
		proto.EntityDTO_VIRTUAL_MACHINE,
	}
	expectedProviderComms := []proto.CommodityDTO_CommodityType{
		proto.CommodityDTO_VMEM,
		proto.CommodityDTO_VCPU,
	}
	var providers []proto.EntityDTO_EntityType
	var providerComms []proto.CommodityDTO_CommodityType
	for key, boughtMap := range appNode.SupportedBoughtComms {
		providers = append(providers, key)
		for comm, _ := range boughtMap {
			providerComms = append(providerComms, comm)
		}
	}
	assert.EqualValues(t, expectedProviders, providers)
	assert.ElementsMatch(t, expectedProviderComms, providerComms)

	assert.EqualValues(t, len(appNode.HostedByProviderType), 1)

	if _, exists := appNode.HostedByProviderType[proto.EntityDTO_VIRTUAL_MACHINE]; !exists {
		assert.Fail(t, "Missing hosted by link")
	}
	hostedRel := appNode.HostedByProviderType[proto.EntityDTO_VIRTUAL_MACHINE]
	assert.True(t, hostedRel == "HOSTING")

	expectedHostedByProviderComms := []proto.CommodityDTO_CommodityType{
		proto.CommodityDTO_VMEM,
		proto.CommodityDTO_VCPU,
		proto.CommodityDTO_TRANSACTION,
		proto.CommodityDTO_RESPONSE_TIME,
	}
	assert.ElementsMatch(t, expectedHostedByProviderComms, appNode.HostedByProviderComms[proto.EntityDTO_VIRTUAL_MACHINE])

	propList := []string{"VM_IP"}
	assert.ElementsMatch(t, propList, appNode.HostedByProviderProps[proto.EntityDTO_VIRTUAL_MACHINE])
}

var SERVICE_NODE string

func TestSupplyChainNodeMissingSoldCommodity(t *testing.T) {
	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   commoditySold:\n" +
			"   - key: key-placeholder \n"

	data := []byte(SERVICE_NODE)
	var sc *conf.SupplyChainConfig
	err := yaml.Unmarshal([]byte(data), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)
}

func TestSupplyChainNodeInvalidType(t *testing.T) {
	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE_ENTITY\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n"

	var sc *conf.SupplyChainConfig
	err := yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)
}

func TestSupplyChainNodeInvalidBought(t *testing.T) {
	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   commodityBought:\n" +
			"     - key:\n" +
			"         templateClass: APPLICATION_COMPONENT\n" +
			"         providerType: LAYERED_OVER\n"

	var sc *conf.SupplyChainConfig
	err := yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)

	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   commodityBought:\n" +
			"     - key:\n" +
			"         templateClass: APP_COMPONENT\n" +
			"         providerType: LAYERED_OVER\n"

	err = yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)
}

func TestSupplyChainNodeInvalidExternalLink(t *testing.T) {
	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   externalLink:\n" +
			"     - key: VIRTUAL_MACHINE\n" +
			"       value:\n" +
			"         buyerRef: APPLICATION_COMPONENT\n" +
			"         sellerRef: VIRTUAL_MACHINE\n" +
			"         relationship: HOSTING\n" +
			"         commodityDefs:\n" +
			"           - type: VCPU\n" +
			"             hasKey: false\n" +
			"         probeEntityPropertyDef:\n" +
			"           - name: VM_IP\n" +
			"             description: IP Address of the VM hosting the discovered node\n" +
			"         externalEntityPropertyDefs:\n" +
			"           - entity: VIRTUAL_MACHINE\n" +
			"             attribute: UsesEndPoints\n" +
			"             propertyHandler:\n" +
			"               methodName: getAddress\n" +
			"               entityType: IP\n" +
			"               directlyApply: false\n"

	var sc *conf.SupplyChainConfig
	err := yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)

	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   externalLink:\n" +
			"     - key: VIRTUAL_MACHINE\n" +
			"       value:\n" +
			"         buyerRef: SERVICE\n" +
			"         sellerRef: VIRTUAL_MACHINE\n" +
			"         relationship: HOSTING\n" +
			"         commodityDefs:\n" +
			"           - type: VCPU\n" +
			"             hasKey: false\n" +
			"         probeEntityPropertyDef:\n" +
			"           - name: VM_IP\n" +
			"             description: IP Address of the VM hosting the discovered node\n"

	err = yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)

	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   externalLink:\n" +
			"     - key: VIRTUAL_MACHINE\n" +
			"       value:\n" +
			"         buyerRef: SERVICE\n" +
			"         sellerRef: VIRTUAL_MACHINE\n" +
			"         relationship: HOSTING\n" +
			"         commodityDefs:\n" +
			"           - type: VCPU\n" +
			"             hasKey: false\n" +
			"         externalEntityPropertyDefs:\n" +
			"           - entity: VIRTUAL_MACHINE\n" +
			"             attribute: UsesEndPoints\n" +
			"             propertyHandler:\n" +
			"               methodName: getAddress\n" +
			"               entityType: IP\n" +
			"               directlyApply: false\n"

	err = yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)

	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   externalLink:\n" +
			"     - key: VIRTUAL_MACHINE\n" +
			"       value:\n" +
			"         buyerRef: SERVICE\n" +
			"         sellerRef: VIRTUAL_MACHINE\n" +
			"         relationship: HOSTING\n" +
			"         probeEntityPropertyDef:\n" +
			"           - name: VM_IP\n" +
			"             description: IP Address of the VM hosting the discovered node\n" +
			"         externalEntityPropertyDefs:\n" +
			"           - entity: VIRTUAL_MACHINE\n" +
			"             attribute: UsesEndPoints\n" +
			"             propertyHandler:\n" +
			"               methodName: getAddress\n" +
			"               entityType: IP\n" +
			"               directlyApply: false\n"

	err = yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)

	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   externalLink:\n" +
			"     - key: VIRTUAL_MACHINE\n" +
			"       value:\n" +
			"         buyerRef: SERVICE\n" +
			"         sellerRef: VIRTUAL_MACHINE\n" +
			"         relationship: HOSTING\n" +
			"         commodityDefs:\n" +
			"           - type: INVALID_COMM\n" +
			"             hasKey: false\n" +
			"         probeEntityPropertyDef:\n" +
			"           - name: VM_IP\n" +
			"             description: IP Address of the VM hosting the discovered node\n" +
			"         externalEntityPropertyDefs:\n" +
			"           - entity: VIRTUAL_MACHINE\n" +
			"             attribute: UsesEndPoints\n" +
			"             propertyHandler:\n" +
			"               methodName: getAddress\n" +
			"               entityType: IP\n" +
			"               directlyApply: false\n"

	err = yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)
}
