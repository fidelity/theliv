/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { SocialUtil } from './social-util';

describe('SocialUtil', () => {
  const socialUtil = new SocialUtil();

  it('should create an instance', () => {
    expect(socialUtil).toBeTruthy();
  });

  // it('should do share when click share icon', () => {
  //   const ev = new Event('click');
  //   spyOn(ev, 'stopPropagation');
  //   const el = fixture.nativeElement.querySelector('.shared-icon');
  //   el.dispatchEvent(ev);
  //   expect(ev.stopPropagation).toHaveBeenCalled();
  // });

  it('should change resource when call thumbs up function', () => {
    const ev = new Event('click');
    spyOn(ev, 'stopPropagation');
    const resource = {
      name: 'test',
      thumbs: false
    };
    socialUtil.isThumbsUp(ev, resource);
    expect(ev.stopPropagation).toHaveBeenCalled();
    expect(resource.thumbs).toBeTrue();
  });

  it('should do share when click share icon', () => {
    const ev = new Event('click');
    spyOn(ev, 'stopPropagation');
    const resource = {
      name: 'test',
      thumbs: false,
      shared: 0
    };
    socialUtil.openShared(ev, resource);
    expect(ev.stopPropagation).toHaveBeenCalled();
    expect(resource.shared).toEqual(1);
  });

  it('should add copy box call copyYamlCode function', () => {
    const spyOnCopy = spyOn(document, 'execCommand');
    socialUtil.copyYamlCode('test');
    expect(spyOnCopy).toHaveBeenCalledWith('copy');
  });
});
