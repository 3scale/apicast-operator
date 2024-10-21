# APIcast-Operator Community Release

[Introduction](#introduction)  
[Main Steps Briefly](#main-steps-briefly)  
[Versioning Strategy](#versioning-strategy)  
[Steps to Perform Community Release in apicast-operator repo](#steps-to-perform-community-release-in-3scale-operator-repo)  
[Prepare release in community-operators-prod repo](#prepare-release-in-community-operators-prod-repo)
[References](#references)


## Introduction
* Community Release is primarily created in parallel with minor or patch Product Releases, but a major community may also be released.
  * Community Release is not required for every productized build or every product patch release.
* Community release can be released after downstream release
* Community Release codebase is defined by Tag in the [apicast-operator](https://github.com/3scale/apicast-operator) repository. 
Tag format as v\<major>.\<minor>.<build#> (for example: v0.6.0, v0.7.1)
* Community Release bundles are defined in `redhat-openshift-ecosystem` in [apicast-community-operator repo](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/apicast-community-operator) . 


## Main Steps Briefly

Below are briefly the main steps for prepare Community Release

### APIcast
- Determine which APIcast image you want to use in the new apicast-operator Community release.
Examples of APIcast images in apicast-operator Community releases:
```asciidoc
apicast-operator    APIcast image
v0.5.0              quay.io/3scale/apicast:v3.11.0
v0.6.0              quay.io/3scale/apicast:v3.12.0
v0.7.1              quay.io/3scale/apicast:v3.13.2
v0.8.0              quay.io/3scale/apicast:3scale-2.14.1-GA
```


### APIcast operator
* Create Community Release Tag - apply tag to the **3scale-2.X-stable** branch to the CommitID of the base Product Release.
  * Tag creation will trigger CircleCI build that will create Operator release Image in this [repo](https://quay.io/repository/3scale/apicast-operator?tab=tags)
* Create a community release development branch in **apicast-operator** repo, based on the Product release CommitID, same CommitID as for Tag,
  or, if Product release tag is not yet available, - based on **3scale-2.X-stable** branch. Do development.
* Create PR from community release development branch to a `3scale-2.X-stable` branch (PR example: https://github.com/3scale/apicast-operator/pull/226)
* Do Development, Test and Merge PR to `3scale-2.X-stable` branch
* Testing - Following Testing will be done before PR merge:
  * Fresh install
  * Upgrade from previous release
* Prepare Community release - create PR in **community-operators-prod** repo (based on manifests from **apicast-operator** repo), 
as in [PR example - apicast-community-operator (0.7.1)](https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/2382)
  * **Merge of this PR will publish Release**.
* Do sanity test of the Release after publishing
  * Fresh install
  * Upgrade from previous release

## Versioning Strategy
Semantic Versioning scheme will be used for Community releases:  MAJOR.MINOR.PATCH.
* MAJOR version: Increments when breaking changes are introduced
* MINOR version: Increments when backward-compatible functionality is added.
* PATCH version: Increments when backward-compatible bug fixes are made.
* APIcast ooperator Community releases Versions examples:  0.6.0,  0.7.1,  0.8.0
* The Version is used in Community releases Tag name

```shell
$ git tag |grep -E "0.6|0.7"     
v0.6.0
v0.7.0
v0.7.1
```

* Tag should be signed. Example of signed tag:
```  
 git show v0.7.1
tag v0.7.1
Tagger: Eguzki Astiz Lezaun <eastizle@redhat.com>
Date:   Tue Mar 14 10:46:34 2023 +0100

v0.7.1
-----BEGIN PGP SIGNATURE-----
xxxxx
xxxxx
-----END PGP SIGNATURE-----

commit 4f3a2e26fa4e97f28cd0d276dc5d46261354afd4 (tag: v0.7.1, tag: show, tag: 3scale-2.13.2-GA, tag: 3scale-2.13.1-GA)
.....
```

## Steps to Perform Community Release in 3scale-operator repo

1. **Release Tag**

- Create Community Release Tag - apply tag to the **3scale-2.X-stable** branch to the CommitID of the base Product Release.
- Tag should be **signed**
- Tag Name will be according to Versioning scheme and have prefix "v", like v0.8.0
  - This tag name convention is required for triggering  of the Release Image creation in repository: 
    https://quay.io/repository/3scale/apicast-operator?tab=tags&tag=v0.8.0

Below is example of Tag creation
  - Look for required CommitID:
```sh
$ cd apicast-operator
$ git switch 3scale-2.14-stable

$ git log
commit f5b5b4b84ce02245cf5787785575546800c02a31 (HEAD -> 3scale-2.14-stable, tag: 3scale-2.14.1-GA, tag: 2.14.1, origin/3scale-2.14-stable)
.....
```
- Create **signed** tag
```sh
$ git tag -s v0.8.0 f5b5b4b84ce02245cf5787785575546800c02a31 -m "Community Release v0.8.0"
```
  _Please note that you will need GPG key to create signed tag (not describe in this doc)_

  - Check tag
```sh
$ git show v0.8.0
tag v0.8.0
Tagger: <user name> <user mail>
Date:   Wed Oct 9 08:47:18 2024 +0300

Community Release v0.8.0
-----BEGIN PGP SIGNATURE-----
XXXXXXX
```

  - Push tag to remote repository
    - Before pushing the Tag - be sure that GPG key that you used to sign the tag is available in your GitHub account / Settings/ GPG keys.
      This will enable GitHub to recognize tags signed with this key as **Verified**.

```shell
$ git push origin v0.8.0
```
  - After tag pushing: check in GitHub that tag is present and has "Verified" badge.

2. **Image creation**
- Image creation process will be triggered after applying the Tag. Image will be created automatically in
  https://quay.io/repository/3scale/apicast-operator repository.
- If the image is not created -
  - check the [apicast-operator CircleCI Pipeline Builds](https://app.circleci.com/pipelines/github/3scale/apicast-operator),
    find your build (by tag), investigate issue, create PR to fix it.
  - after fixing and PR merge - delete the tag and create a signed tag again. Recheck CircleCI Pipeline ...

3. **Community Release Development Branch**

* Create a Community release development branch from **3scale-2.X-stable** branch **based on Product Release CommitID**
  **Note**: same CommitID as for Community Release Tag
```
$ cd apicast-operator
$ git switch 3scale-2.14-stable
Switched to branch '3scale-2.14-stable'
$ git log
commit f5b5b4b84ce02245cf5787785575546800c02a31 (HEAD -> 3scale-2.14-stable, tag: 3scale-2.14.1-GA, tag: 2.14.1, origin/3scale-2.14-stable)
```
```
$ git checkout -b apicast-community-v0.8.0 f5b5b4b84ce02245cf5787785575546800c02a31
```

4. **APIcast Components Images - APIcast**
* Community release can use its own APIcast image or use existing build of APIcast. 
  You can search [APIcast repository](https://github.com/3scale/APIcast) for APIcast version and Tag to be used
  in Community APIcast operator. For example: tag: 3scale-2.14.1-GA, tag: v3.14.0 ...


5. **Create PR for apicast-operator Community release development branch**
   * Work on PR to have E2E tests passed
   * Test fresh Install and Upgrade
   * Merge PR to **3scale-2.X-stable** branch
   * [Example of the PR, for Community release v0.8.0](https://github.com/3scale/apicast-operator/pull/228)
  

6. **Testing**

   * Testing of `apicast-operator` release development PR must be completed before open PR in `community-operators-prod` repo.
   * Testing must be done for `Fresh Install` and `Upgrade`.

## Prepare release in community-operators-prod repo
See Documentation in [Pull Requests in community operators project](https://github.com/operator-framework/community-operators/blob/master/docs/contributing-via-pr.md)

**IMPORTANT. All previouse steps, including Testing of apicast-operator (Fresh install and Upgrade) must be completed 
  before opening a PR in the community-operators-prod repo**

* Fork & Clone community-operators-prod repo if you don't have it yet
* Get latest changes

```
$ git clone git@github.com:redhat-openshift-ecosystem/community-operators-prod.git
$ cd community-operators-prod

Example for myfork:
$ git remote -v
myfork        git@github.com:valerymo/community-operators-prod.git (fetch)
myfork        git@github.com:valerymo/community-operators-prod.git (push)
origin        git@github.com:redhat-openshift-ecosystem/community-operators-prod.git (fetch)
origin        git@github.com:redhat-openshift-ecosystem/community-operators-prod.git (push)
```

* Create branch in community-operators-prod repo
```
$ git checkout -b apicast-community-operator-v0.8.0
```

* Create a new release folder (similar to previous release)
```
community-operators-prod/operators/apicast-community-operator/0.8.0
```

* Copy manifests from apicast-operator Community release to community-operators-prod, as for example From  [apicast-operator PR](https://github.com/3scale/apicast-operator/pull/226) To [community-operators-prod/operators/apicast-community-operator/0.8.0](community-operators-prod/operators/apicast-community-operator/0.8.0)

* Update CSV. These are things that need pay attention:
  *  apicast-operator containerImage
  *  apicast image
  *  CVS version
  *  CSV replaces
  *  CSV name
  *  CSV description
  *  CSV urls
  *  Annotations
  * * Update metadata/annotations  and bundle.Dockerfile
  

* Compare apicast-community-operator bundle with previous version, and with apicast-operator
* Finally you will have update release bundle structure.
* Commit your changes and Create PR 
  * Do Signed-off  - git commit -s … , It’s required for PR test Pipeline

```
[community-operators-prod] (apicast-community-operator-v0.8.0)$ git log
commit xxxxxxxxxxxx (HEAD -> apicast-community-operator-v0.8.0)
Author: xxx xxx <xxx@xxx.com>
Date:   xxx
apicast-community-operator release v0.8.0
Signed-off-by: xxx xxx <xxx@xxx.com>
```

```
$ git push -u myfork apicast-community-operator-v0.8.0
To github.com:valerymo/community-operators-prod.git
....
```

**IMPORTANT. Merging of community-operators-prod PR will trigger Release creation/publishing**

* Check and confirm all questions in PR description
* Merge
* Test of created release in OSD cluster / OperatorHub
________________


## References

* [Community operators project documentation](https://github.com/operator-framework/community-operators/blob/master/docs/contributing-via-pr.md)
* [Community operators repository](https://github.com/k8s-operatorhub/community-operators)
* [K8S PR example](https://github.com/k8s-operatorhub/community-operators/pull/2472)
* [OCP PR example](https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/2382)
* *Check for the latest guidelines about what it needs to be done; they change from time to time*
  * *Community operator release process moved* [here - operator-release-process](https://github.com/operator-framework/community-operators/blob/master/docs/operator-release-process.md)
