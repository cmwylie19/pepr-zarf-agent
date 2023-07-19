import { Capability, a, Log } from "pepr";
import { K8sAPI } from "./kubernetes-api";
import { InitSecrets } from "./secrets/initSecrets";
import { InitSecretsReady } from "./helpers";
import { TransformerAPI } from "./transformer-api";
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
      }
    }

    // Create a imagePullSecret in Pod namespace
    if (pod.HasLabel("zarf-agent") || pod.HasLabel("zarf.dev/agent")) {
      Log.info("Pod has ignore labels. Skipping.");
    } else {
      Log.info("Pod does not have ignore labels. Continuing.");

      try {
        // check if imagePullSecrets exist
        if (
          pod.Raw?.spec?.imagePullSecrets !== undefined ||
          pod.Raw?.spec?.imagePullSecrets !== null
        ) {
          pod.Raw.spec.imagePullSecrets = [];
        }
        // add imagePullSecret to pod
        pod.Raw?.spec?.imagePullSecrets?.push({
          name: _initSecrets.privateRegistrySecretName,
        });
      } catch (err) {
        Log.error("Could not add imagePullSecret to pod", err);
      }

      try {
        // transform all containers in pod
        _transformer.run().then(() => {
          _transformer.transformAllContainers(
            pod,
            _initSecrets.zarfStateSecret.registryInfo.address
          );
        });
      // pod.Merge() deepPartial from WASM

        // add zarf-agent label to pod to be ignored next time
        pod.SetAnnotation("zarg-agent/dev", "patched");
      } catch (err) {
        Log.error("Could not patch images of pod", err);
      }
    }
  });
