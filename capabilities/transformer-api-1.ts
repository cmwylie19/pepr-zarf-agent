import { Log, PeprRequest, a } from "pepr";
import { ImageTransformClient } from "./lib/images/image_grpc_pb";
import { TransformRequest } from "./lib/images/image_pb";
import { credentials } from "@grpc/grpc-js";
import "wasm_exec";
const H0STNAME: string = "transformer.pepr-system.svc.cluster.local:50051";

export class TransformerAPI {
  client: ImageTransformClient;

  constructor() {
    this.client = new ImageTransformClient(
      H0STNAME,
      credentials.createInsecure()
    );
  }

  async transformAllContainers(
    pod: PeprRequest<a.Pod>,
    address: string
  ): Promise<void> {
    if (pod.Raw?.spec?.initContainers !== undefined) {
      await Promise.all(
        pod.Raw.spec.initContainers.map(async container => {
          try {
            container.image = await this.imageTransformHost(
              address,
              container.image
            );
          } catch (err) {
            Log.error("Error calling imageTransformHost", err.toString());
          }
        })
      );
    }

    if (pod.Raw?.spec?.ephemeralContainers !== undefined) {
      await Promise.all(
        pod.Raw.spec.ephemeralContainers.map(async container => {
          try {
            container.image = await this.imageTransformHost(
              address,
              container.image
            );
          } catch (err) {
            Log.error("Error calling imageTransformHost", err.toString());
          }
        })
      );
    }

    await Promise.all(
      pod.Raw.spec.containers.map(async container => {
        try {
          container.image = await this.imageTransformHost(
            address,
            container.image
          );
          Log.info(
            "Transformed image: " +
              JSON.stringify(pod.Raw.spec.containers, undefined, 2)
          );
        } catch (err) {
          Log.error("Error calling imageTransformHost", err.toString());
        }
      })
    );
  }

  async imageTransformHost(
    targetHost: string,
    srcReference: string
  ): Promise<string> {
    const request = new TransformRequest();
    request.setTargethost(targetHost);
    request.setSrcreference(srcReference);

    return new Promise<string>((resolve, reject) => {
      this.client.imageTransformHost(request, (err, response) => {
        if (err) {
          Log.error("Error calling imageTransformHost", err.toString());
          reject(err);
        } else {
          Log.info("Transformed image: " + response.getTransformedimage());
          resolve(response.getTransformedimage());
        }
      });
    });
  }

  async imageTransformHostWithoutChecksum(
    targetHost: string,
    srcReference: string
  ): Promise<string> {
    const request = new TransformRequest();
    request.setTargethost(targetHost);
    request.setSrcreference(srcReference);

    return new Promise<string>((resolve, reject) => {
      this.client.imageTransformHostWithoutChecksum(
        request,
        (err, response) => {
          if (err) {
            Log.error(
              "Error calling imageTransformHostWithoutChecksum",
              err.toString()
            );
            reject(err);
          } else {
            resolve(response.getTransformedimage());
          }
        }
      );
    });
  }
}
