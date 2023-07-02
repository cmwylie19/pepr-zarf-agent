import { Log, PeprRequest, a } from "pepr";
import { ImageTransformClient } from "./lib/images/image_grpc_pb";
import { TransformRequest } from "./lib/images/image_pb";
import { credentials } from "@grpc/grpc-js";

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
      Promise.all(
        pod.Raw.spec.initContainers.map(
          async container =>
            await this.imageTransformHost(address, container.image)
        )
      );
    }

    if (pod.Raw?.spec?.ephemeralContainers !== undefined) {
      Promise.all(
        pod.Raw.spec.ephemeralContainers.map(
          async container =>
            await this.imageTransformHost(address, container.image)
        )
      );
    }
    if (pod.Raw?.spec?.containers !== undefined) {
      Promise.all(
        pod.Raw.spec.containers.map(
          async container =>
            await this.imageTransformHost(address, container.image)
        )
      );
    }
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