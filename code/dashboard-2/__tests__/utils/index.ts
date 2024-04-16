import { isValidateClusterName } from '../../src/util'


describe('isValidateClusterName', () => {
  it('should return true for a valid cluster name', () => {
    expect(isValidateClusterName('ml')).toBe(true);
  })

  it('should return false for an invalid cluster name', () => {
    expect(isValidateClusterName('invalidCluster')).toBe(false);
  })

  it('should return false for an undefined cluster name', () => {
    expect(isValidateClusterName(undefined)).toBe(false);
  })

  it('should return false for null cluster name', () => {
    expect(isValidateClusterName(null)).toBe(false);
  })

  it('should return false for an empty cluster name', () => {
    expect(isValidateClusterName("")).toBe(false);
  })

  it('should return false for an empty cluster name', () => {
    expect(isValidateClusterName()).toBe(false);
  })
});
