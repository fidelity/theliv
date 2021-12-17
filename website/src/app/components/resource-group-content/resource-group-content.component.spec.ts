import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BehaviorSubject, of } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { KubernetesService } from '../../services/kubernetes.service';
import { SocialUtil } from '../../shared/util/social-util';
import { ResourceGroupContentComponent } from './resource-group-content.component';

class MockedKubernetesService {
  public resourceList$: BehaviorSubject<any> = new BehaviorSubject<any>([
    {
      name: 'crash-deployment',
      rootCause: {},
      topResourceType: 'Deployment',
      resources: [
        {
          name: 'crash-pod',
          type: 'Pod',
          issue: {
            name: 'CrashLoopBackOff',
            description: 'this is the issue discription'
          },
          metadata: {
            spec: {
              nodeName: 'ip-10-000-111-001.ec2.internal',
              containers: [
                {
                  image: 'test-image'
                }
              ]
            },
            status: {
              podIP: 'aa.bbb.ccc.ddd'
            }
          }
        }
      ],
      id: '111'
    },
    {
      name: 'image-failed-deployment',
      rootCause: {},
      topResourceType: 'Deployment',
      resources: [
        {
          name: 'failed-pod',
          type: 'Pod',
          issue: {
            name: 'ImagePullBackoff'
          }
        }
      ],
      id: '222'
    }
  ]);
  public selectedClusters$: BehaviorSubject<any> = new BehaviorSubject<any>('');
  public selectedNs$: BehaviorSubject<any> = new BehaviorSubject<any>('');
}

const MockactiveRoute = {
  params: of({ id: '111' })
};

describe('ResourceGroupContentComponent', () => {
  let component: ResourceGroupContentComponent;
  let fixture: ComponentFixture<ResourceGroupContentComponent>;
  let kubernetesService: KubernetesService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ResourceGroupContentComponent],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: MockactiveRoute
        },
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
    fixture = TestBed.createComponent(ResourceGroupContentComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should get pod resrouces', () => {
    expect(component.podResource.length).toBeGreaterThan(0);
  });


  // it('should show popup content when call showPopupItem function', () => {
  //   component.showPopupItem();
  //   expect(component.gridToggle).toBeTruthy();
  // });

});
