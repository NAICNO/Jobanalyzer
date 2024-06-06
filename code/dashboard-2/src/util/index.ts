import { CLUSTER_INFO } from '../Constants.ts'

export const isValidateClusterName = (clusterName?: string | null): boolean => {
  return clusterName ? !!CLUSTER_INFO[clusterName] : false;
}

export const breakText = (text: string): string => {
    // Don't break at spaces that exist, but at commas.  Generally this has the effect of
    // keeping duration and timestamp fields together and breaking the command field apart.
    //
    // TODO: This is not ideal, b/c it breaks within node name ranges, we can fix that.
    return text.replaceAll(" ", "\xA0").replaceAll(",", ", ")
}
