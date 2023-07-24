import { Capability, a, Log } from "pepr";
import { K8sAPI } from "./kubernetes-api";
import { InitSecrets } from "./secrets/initSecrets";
import { InitSecretsReady } from "./helpers";
import { TransformerAPI } from "./transformer-api";
import { log } from "@grpc/grpc-js/build/src/logging";
/**
 *  The ZarfAgent capability handles pod mutations for Zarf.
 */
export const ZarfAgent = new Capability({
  name: "zarf-agent",
  description: "A mutating webhook for Zarf.",
  namespaces: [], // all namespaces
});

// Use the 'When' function to create a new Capability Action
const { When } = ZarfAgent;

/**
 * ---------------------------------------------------------------------------------------------------
 * Initialize InitSecrets & TransformerAPI
 */
const _initSecrets = new InitSecrets(new K8sAPI());
const _transformer = new TransformerAPI();

/**
 * ---------------------------------------------------------------------------------------------------
 *                                   CAPABILITY ACTION (Pod)                                   *
 * ---------------------------------------------------------------------------------------------------
 *
 * This Capability Action fetches the `zarf-state` and `private-registry` secrets when
 * a pod is created, saves them to state, and deploys the `private-registry` secret to the
 * pod namespace.
 */
When(a.ConfigMap)
  .IsCreated()
  .Then(() => {
    try {
      Log.info(
        "Private Registry Secret",
        JSON.stringify(
          _initSecrets.privateRegistrySecretData[".dockerconfigjson"],
          undefined,
          2
        )
      );
      Log.info(
        "Zarf State Secret",
        JSON.stringify(_initSecrets.zarfStateSecretData["state"], undefined, 2)
      );
    } catch (err) {
      Log.error(
        "Could not fetch secrets because pod has not been created",
        err
      );
    }
  });

When(a.Pod)
  .IsCreatedOrUpdated()
  .Then(async pod => {
    // Turn up logging
    Log.SetLogLevel("debug");

    // If InitSecrets do not exist, create them
    if (!InitSecretsReady(_initSecrets)) {
      try {
        await _initSecrets.getZarfStateSecret();
        await _initSecrets.getZarfPrivateRegistrySecret();

        Log.info(`InitSecrets initialized. 💯`);
      } catch (err) {
        Log.error("Secrets in zarf namespace do not exist", err);
        return
      }
    }

 
      try {
        // transform all containers in pod
        await _transformer.run();
       
          let result = _transformer.transformPod(
            pod.Raw,
            pod.Request,
            _initSecrets.privateRegistrySecretName,
            _initSecrets.zarfStateSecret.registryInfo.address
          );
          Log.info(`Result. ${JSON.stringify(result, undefined, 2)}`);
      

    } catch (err) {
      Log.error("Error transforming pod", err);
    }
    console.log("pod", JSON.stringify(pod, undefined, 2))
  });
