package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	dialTimeout = 5 * time.Second
	config      *clientv3.Config
	client      *clientv3.Client
)

const (
	THELIV_CONFIG_KEY  string = "/theliv/config"
	DATADOG_CONFIG_KEY string = "/theliv/config/datadog"
	THELIV_AUTH_KEY    string = "/theliv/config/authconf"
	EKS_CLUSTERS_KEY   string = "/theliv/clusters/eks"
)

// Init client config, could be called only once, before any other functions
func InitClientConfig(ca string, cert string, key string, endpoints []string) {
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Panicf("Failed to load cert pair for etcd cluster, %v", err)
	}

	caData, err := ioutil.ReadFile(ca)
	if err != nil {
		log.Panicf("Failed to load ca cert for etcd cluster, %v", err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	_tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{c},
		RootCAs:      pool,
	}

	config = &clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		TLS:         _tlsConfig,
	}
	client = newClient()
}

func newClient() *clientv3.Client {
	if config == nil {
		log.Panicf("Etcd client config is not initialized yet!")
	}
	client, err := clientv3.New(*config)
	if err != nil {
		log.Panicf("Failed to init etcd client, error : %v", client)
	}
	return client
}

// Put KV to etcd
func PutStr(key, value string) error {
	// client := newClient()
	// defer client.Close()

	_, err := client.Put(context.TODO(), key, value)
	log.Printf("Failed to put %v to etcd\n", key)
	return err
}

// Marshall the value (struct) and put to etcd
func Put(key string, value interface{}) error {
	c, err := json.Marshal(value)
	if err != nil {
		log.Printf("Failed to marshall %v\n", value)
		return err
	}
	return PutStr(key, string(c))
}

// Get keys only with prefix
func GetKeys(prefix string) ([]string, error) {
	// client := newClient()
	// defer client.Close()

	res, err := client.Get(context.TODO(), prefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		log.Printf("Failed to get %v with keys only and prefix, error is %v\n", prefix, err)
		return nil, err
	}
	keys := make([]string, 0)
	if res.Kvs == nil {
		log.Printf("No keys found for prefix %v\n", prefix)
		return keys, nil
	}
	for _, k := range res.Kvs {
		keys = append(keys, string(k.Key))
	}
	return keys, nil
}

// Get object (struct) from etcd, assume all the data in etcd should be in json format
// value should be a pointer
func GetObject(key string, value interface{}) error {
	// client := newClient()
	// defer client.Close()
	res, err := client.Get(context.TODO(), key)
	if err != nil {
		log.Printf("Failed to get %v, error is %v\n", key, err)
		return err
	}
	if l := len(res.Kvs); l != 1 {
		log.Printf("Get %v keys from etcd\n", l)
		return err
	}
	//assume all the value should be in json format
	err = json.Unmarshal(res.Kvs[0].Value, value)
	if err != nil {
		log.Printf("Failed to unmarshall value to %v\n", value)
	}
	return err
}

// Get content from key directly
func Get(key string) ([]byte, error) {
	// client := newClient()
	// defer client.Close()
	res, err := client.Get(context.TODO(), key)
	if err != nil {
		log.Printf("Failed to get %v, error is %v\n", key, err)
		return nil, err
	}
	if l := len(res.Kvs); l != 1 {
		log.Printf("Get %v keys from etcd\n", l)
		return nil, err
	}
	return res.Kvs[0].Value, nil
}

// Get the json value of the key, also the sub paths
// Adding the sub values to parent, ASSUME the type of sub is []byte
func GetObjectWithSub(key string, obj interface{}) error {
	// client := newClient()
	// defer client.Close()
	res, err := client.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		log.Printf("Failed to get prefix %v, error is %v\n", key, err)
		return err
	}
	firstkey := key + "/"
	for _, value := range res.Kvs {
		// If the key present in etcd, unmarshal the content to obj
		if key == string(value.Key) {

			err := json.Unmarshal(value.Value, obj)
			if err != nil {
				log.Printf("Failed to unmarshall value to obj, the etcd key is: %v\n", firstkey)
			}
		} else {
			k := string(value.Key)
			sub := strings.Replace(k, firstkey, "", -1)
			if field, ok := getFieldByTag(obj, "json", sub); ok {
				setStructFieldValue(field, value.Value)
			}
		}
	}
	return nil
}

func setStructFieldValue(field *reflect.Value, value []byte) {
	switch field.Kind() {
	case reflect.Struct:
		// for struct type, create new struct and unmarshal
		v := reflect.New(field.Type()).Elem()
		err := json.Unmarshal(value, v.Addr().Interface())
		if err != nil {
			log.Printf("Failed to unmarshall value to field %v \n", field.Type().Name())
		}
		field.Set(v)
	case reflect.Slice:
		//assume byte slice
		field.SetBytes(value)
	}
}

// Get both keys and values start with prefix
func GetWithPrefix(pre string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	// client := newClient()
	// defer client.Close()
	res, err := client.Get(context.TODO(), pre, clientv3.WithPrefix())
	if err != nil {
		log.Printf("Failed to call GetWithPrefix, prefix %v, error is %v\n", pre, err)
		return result, err
	}
	for _, kv := range res.Kvs {
		result[string(kv.Key)] = kv.Value
	}
	return result, nil
}

func getFieldByTag(obj interface{}, tag, name string) (*reflect.Value, bool) {
	instance := reflect.ValueOf(obj).Elem()

	for _, f := range reflect.VisibleFields(instance.Type()) {
		if match := f.Tag.Get(tag) == name; match {
			field := instance.FieldByIndex(f.Index)
			return &field, true
		}
	}
	return nil, false
}
