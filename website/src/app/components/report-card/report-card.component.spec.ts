/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BehaviorSubject } from 'rxjs';
import { KubernetesService } from '../../services/kubernetes.service';
import { SocialUtil } from '../../shared/util/social-util';
import { ReportCardComponent } from './report-card.component';

class MockedKubernetesService {
  public resourceList$: BehaviorSubject<any> = new BehaviorSubject<any>([
    {
      name: 'image-failed-deployment',
      rootCause: {},
      topResourceType: 'Deployment',
      resources: [
        {
          name: 'test-pod',
          type: 'Pod',
          issue: {
            name: 'CrashLoopBackOff'
          }
        }
      ],
      id: '3199429993'
    }
  ]);
  public selectedClusters$: BehaviorSubject<any> = new BehaviorSubject<any>('');
  public selectedNs$: BehaviorSubject<any> = new BehaviorSubject<any>('');
}

describe('ReportCardComponent', () => {
  let component: ReportCardComponent;
  let fixture: ComponentFixture<ReportCardComponent>;
  let kubernetesService: KubernetesService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ReportCardComponent ],
      providers: [
        {
            provide: KubernetesService,
            useClass: MockedKubernetesService
        },
        SocialUtil
      ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    kubernetesService = TestBed.inject(KubernetesService);
    fixture = TestBed.createComponent(ReportCardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  // it('should show popup content when call showPopupItem function', () => {
  //   component.showPopupItem(0);
  //   expect(component.visibleIndex).toEqual(0);
  // });

  // it('should close popup content when call showPopupItem function with same index', () => {
  //   component.visibleIndex = 0;
  //   component.showPopupItem(0);
  //   expect(component.visibleIndex).toEqual(-1);
  // });

});
