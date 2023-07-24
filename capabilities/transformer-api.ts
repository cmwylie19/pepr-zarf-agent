import { Log, PeprRequest, a } from "pepr";
import { readFileSync } from "fs";
import "./wasm_exec.js";

export class TransformerAPI {
  private go: any;
  private instance: WebAssembly.Instance;

  transform(
    pod: string,
    request: string,
    imagePullSecretName: string,
    targetHost: string
  ): string {
    // @ts-ignore
    return zarfTransform.podTransform(
      pod,
      request,
      imagePullSecretName,
      targetHost
    );
  }
  // this create global zarfTransform object
  private async instantiateWebAssembly(): Promise<void> {
    return WebAssembly.instantiate(
      readFileSync("capabilities/main.wasm"),
      this.go.importObject
    ).then(wasmModule => {
      this.instance = this.go.run(wasmModule.instance);
    });
  }

  public async run(): Promise<void> {
    this.go = new globalThis.Go();
    try {
      await this.instantiateWebAssembly();
    } catch (err) {
      Log.error("Error instantiating wasm module", err.toString());
      return;
    }
  }
  transformPod(
    pod: any,
    request: any,
    imagePullSecretName: string,
    targetHost: string
  ): string {
    let transformedPod: string;
    if (!this.instance) {
      throw new Error("WebAssembly module not loaded or initialized.");
    }
    try {
      transformedPod = this.transform(
        JSON.stringify(pod),
        JSON.stringify(request),
        imagePullSecretName,
        targetHost
      );
    } catch (err) {
      Log.error("Error calling imageTransformHost", err);
    }
    return transformedPod;
  }
}
